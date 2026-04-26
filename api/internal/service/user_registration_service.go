package service

import (
	"context"
	"encoding/json"
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
	mailer                      mail.Mailer
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
	mailer mail.Mailer,
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
		mailer:                      mailer,
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
		if err := s.mailer.SendUserAlreadyRegisteredMail(ctx, input.Email, input.Language); err != nil {
			return nil, err
		}
		return &CreateUserRegistrationOutput{Code: i18n.CodeUserRegistrationRequestCreated}, nil
	}

	if req != nil && req.LastSentAt != nil {
		resendAvailableAt := req.LastSentAt.Add(time.Duration(s.config.RegistrationResendIntervalMinutes()) * time.Minute)
		if now.Before(resendAvailableAt) {
			return &CreateUserRegistrationOutput{Code: i18n.CodeUserRegistrationRequestCreated}, nil
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

		// ★ Outbox登録（TX内）
		outboxID, err := s.uuidGenerator.NewV7()
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]string{
			"email": input.Email,
			"token": plainToken,
			"lang":  input.Language,
		})
		if err != nil {
			return err
		}

		if err := s.mailOutboxRepo.Create(txCtx, &model.MailOutbox{
			ID:            outboxID,
			MailType:      "user_registration",
			ToEmail:       input.Email,
			Payload:       string(payload),
			Status:        "pending",
			RetryCount:    0,
			NextAttemptAt: now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// ※ ここでは従来通り送信（後でdispatcherに完全移行）
	registerURL := s.registrationURLBuilder.Build(plainToken)

	if err := s.mailer.SendUserRegistrationMail(ctx, mail.UserRegistrationMail{
		To:             input.Email,
		URL:            registerURL,
		Lang:           input.Language,
		ExpiresMinutes: s.config.RegistrationTokenExpiresMinutes(),
	}); err != nil {
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
