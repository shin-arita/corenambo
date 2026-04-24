package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"app-api/internal/app_error"
	"app-api/internal/clock"
	"app-api/internal/config"
	"app-api/internal/i18n"
	"app-api/internal/mail"
	"app-api/internal/model"
	"app-api/internal/registrationurl"
	"app-api/internal/repository"
	"app-api/internal/token"
	"app-api/internal/uuid"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type UserRegistrationService interface {
	Create(ctx context.Context, input CreateUserRegistrationInput) (*CreateUserRegistrationOutput, error)
}

type CreateUserRegistrationInput struct {
	Email             string
	EmailConfirmation string
	Language          string
}

type CreateUserRegistrationOutput struct {
	Code string
}

type userRegistrationService struct {
	userRegistrationRequestRepo repository.UserRegistrationRequestRepository
	txManager                   repository.TxManager
	tokenGenerator              token.Generator
	tokenHasher                 token.Hasher
	uuidGenerator               uuid.Generator
	clock                       clock.Clock
	mailer                      mail.Mailer
	registrationURLBuilder      registrationurl.Builder
	config                      config.Config
}

func NewUserRegistrationService(
	userRegistrationRequestRepo repository.UserRegistrationRequestRepository,
	txManager repository.TxManager,
	tokenGenerator token.Generator,
	tokenHasher token.Hasher,
	uuidGenerator uuid.Generator,
	clock clock.Clock,
	mailer mail.Mailer,
	registrationURLBuilder registrationurl.Builder,
	config config.Config,
) UserRegistrationService {
	return &userRegistrationService{
		userRegistrationRequestRepo: userRegistrationRequestRepo,
		txManager:                   txManager,
		tokenGenerator:              tokenGenerator,
		tokenHasher:                 tokenHasher,
		uuidGenerator:               uuidGenerator,
		clock:                       clock,
		mailer:                      mailer,
		registrationURLBuilder:      registrationURLBuilder,
		config:                      config,
	}
}

func (s *userRegistrationService) Create(
	ctx context.Context,
	input CreateUserRegistrationInput,
) (*CreateUserRegistrationOutput, error) {
	if err := validateCreateInput(input); err != nil {
		return nil, err
	}

	input.Email = strings.TrimSpace(input.Email)
	input.EmailConfirmation = strings.TrimSpace(input.EmailConfirmation)

	now := s.clock.Now()

	req, err := s.userRegistrationRequestRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if req != nil && req.VerifiedAt != nil {
		return nil, app_error.NewConflict(i18n.CodeUserAlreadyRegistered)
	}

	plainToken, err := s.tokenGenerator.Generate()
	if err != nil {
		return nil, err
	}

	tokenHash, err := s.tokenHasher.Hash(plainToken)
	if err != nil {
		return nil, err
	}

	expiresAt := now.Add(time.Duration(s.config.RegistrationTokenExpiresMinutes()) * time.Minute)

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if req == nil {
			id, err := s.uuidGenerator.NewV7()
			if err != nil {
				return err
			}

			newReq := &model.UserRegistrationRequest{
				ID:         id,
				Email:      input.Email,
				TokenHash:  tokenHash,
				ExpiresAt:  expiresAt,
				VerifiedAt: nil,
				CreatedAt:  now,
			}

			if err := s.userRegistrationRequestRepo.Create(txCtx, newReq); err != nil {
				return err
			}
		} else {
			req.TokenHash = tokenHash
			req.ExpiresAt = expiresAt

			if err := s.userRegistrationRequestRepo.UpdateToken(txCtx, req); err != nil {
				return err
			}
		}

		registerURL := s.registrationURLBuilder.Build(plainToken)

		if err := s.mailer.SendUserRegistrationMail(txCtx, mail.UserRegistrationMail{
			To:   input.Email,
			URL:  registerURL,
			Lang: input.Language,
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CreateUserRegistrationOutput{
		Code: i18n.CodeUserRegistrationRequestCreated,
	}, nil
}

func validateCreateInput(input CreateUserRegistrationInput) error {
	fieldErrors := map[string][]app_error.FieldError{}

	email := strings.TrimSpace(input.Email)
	emailConfirmation := strings.TrimSpace(input.EmailConfirmation)

	if email == "" {
		fieldErrors["email"] = append(fieldErrors["email"], app_error.FieldError{
			Code: i18n.CodeEmailRequired,
		})
	} else if !emailRegex.MatchString(email) {
		fieldErrors["email"] = append(fieldErrors["email"], app_error.FieldError{
			Code: i18n.CodeEmailFormatInvalid,
		})
	}

	if emailConfirmation == "" {
		fieldErrors["email_confirmation"] = append(fieldErrors["email_confirmation"], app_error.FieldError{
			Code: i18n.CodeEmailConfirmationRequired,
		})
	} else if email != "" && email != emailConfirmation {
		fieldErrors["email_confirmation"] = append(fieldErrors["email_confirmation"], app_error.FieldError{
			Code: i18n.CodeEmailConfirmationNotMatch,
		})
	}

	if len(fieldErrors) > 0 {
		return app_error.NewValidation(fieldErrors)
	}

	return nil
}
