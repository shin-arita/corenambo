package service

import (
	"context"
	"encoding/json"
	"net/mail"
	"strings"
	"time"

	"app-api/internal/app_error"
	"app-api/internal/clock"
	"app-api/internal/config"
	"app-api/internal/i18n"
	"app-api/internal/model"
	"app-api/internal/registrationurl"
	"app-api/internal/repository"
	"app-api/internal/token"
	"app-api/internal/uuid"
)

type UserRegistrationService interface {
	Create(ctx context.Context, input CreateUserRegistrationInput) (*CreateUserRegistrationOutput, error)
	Verify(ctx context.Context, input VerifyUserRegistrationInput) (*VerifyUserRegistrationOutput, error)
}

type CreateUserRegistrationInput struct {
	Email             string
	EmailConfirmation string
	Language          string
}

type CreateUserRegistrationOutput struct {
	Code string
}

type VerifyUserRegistrationInput struct {
	Token string
}

type VerifyUserRegistrationOutput struct {
	Code string
}

type userRegistrationService struct {
	userRegistrationRequestRepo repository.UserRegistrationRequestRepository
	mailOutboxRepo              repository.MailOutboxRepository
	txManager                   repository.TxManager
	tokenGenerator              token.Generator
	tokenHasher                 token.Hasher
	uuidGenerator               uuid.Generator
	clock                       clock.Clock
	registrationURLBuilder      registrationurl.Builder
	config                      config.Config
}

func NewUserRegistrationService(
	userRegistrationRequestRepo repository.UserRegistrationRequestRepository,
	mailOutboxRepo repository.MailOutboxRepository,
	txManager repository.TxManager,
	tokenGenerator token.Generator,
	tokenHasher token.Hasher,
	uuidGenerator uuid.Generator,
	clock clock.Clock,
	registrationURLBuilder registrationurl.Builder,
	config config.Config,
) UserRegistrationService {
	return &userRegistrationService{
		userRegistrationRequestRepo: userRegistrationRequestRepo,
		mailOutboxRepo:              mailOutboxRepo,
		txManager:                   txManager,
		tokenGenerator:              tokenGenerator,
		tokenHasher:                 tokenHasher,
		uuidGenerator:               uuidGenerator,
		clock:                       clock,
		registrationURLBuilder:      registrationURLBuilder,
		config:                      config,
	}
}

func validateCreateInput(input CreateUserRegistrationInput) error {
	fieldErrors := map[string][]app_error.FieldError{}

	email := strings.TrimSpace(input.Email)
	emailConfirmation := strings.TrimSpace(input.EmailConfirmation)

	if email == "" {
		fieldErrors["email"] = append(fieldErrors["email"], app_error.FieldError{
			Code: i18n.CodeEmailRequired,
		})
	} else if _, err := mail.ParseAddress(email); err != nil {
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
		return &CreateUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationRequestCreated,
		}, nil
	}

	if req != nil && req.LastSentAt != nil {
		resendAvailableAt := req.LastSentAt.Add(
			time.Duration(s.config.RegistrationResendIntervalMinutes()) * time.Minute,
		)
		if now.Before(resendAvailableAt) {
			return &CreateUserRegistrationOutput{
				Code: i18n.CodeUserRegistrationRequestCreated,
			}, nil
		}
	}

	plainToken, err := s.tokenGenerator.Generate()
	if err != nil {
		return nil, err
	}

	tokenHash, err := s.tokenHasher.Hash(plainToken)
	if err != nil {
		return nil, err
	}

	expiresAt := now.Add(
		time.Duration(s.config.RegistrationTokenExpiresMinutes()) * time.Minute,
	)

	registrationURL := s.registrationURLBuilder.Build(plainToken)

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
				LastSentAt: &now,
				CreatedAt:  now,
			}

			if err := s.userRegistrationRequestRepo.Create(txCtx, newReq); err != nil {
				return err
			}
		} else {
			req.TokenHash = tokenHash
			req.ExpiresAt = expiresAt
			req.LastSentAt = &now

			if err := s.userRegistrationRequestRepo.UpdateToken(txCtx, req); err != nil {
				return err
			}
		}

		outboxID, err := s.uuidGenerator.NewV7()
		if err != nil {
			return err
		}

		payload, _ := json.Marshal(map[string]string{
			"email": input.Email,
			"url":   registrationURL,
			"lang":  input.Language,
		})

		return s.mailOutboxRepo.Create(txCtx, &model.MailOutbox{
			ID:            outboxID,
			MailType:      "user_registration",
			ToEmail:       input.Email,
			Payload:       string(payload),
			Status:        "pending",
			RetryCount:    0,
			NextAttemptAt: now,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	})

	if err != nil {
		return nil, err
	}

	return &CreateUserRegistrationOutput{
		Code: i18n.CodeUserRegistrationRequestCreated,
	}, nil
}

func (s *userRegistrationService) Verify(
	ctx context.Context,
	input VerifyUserRegistrationInput,
) (*VerifyUserRegistrationOutput, error) {

	return &VerifyUserRegistrationOutput{
		Code: "OK",
	}, nil
}
