package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

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

var jsonMarshal = json.Marshal

var bcryptGenerate = bcrypt.GenerateFromPassword

var hashPassword = func(password string) (string, error) {
	b, err := bcryptGenerate([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

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
	Code           string
	ExpiresMinutes int
}

type VerifyUserRegistrationInput struct {
	Token                string
	DisplayName          string
	Password             string
	PasswordConfirmation string
	AgreedToTerms        bool
}

type VerifyUserRegistrationOutput struct {
	Code string
}

type userRegistrationService struct {
	userRegistrationRequestRepo repository.UserRegistrationRequestRepository
	userRepo                    repository.UserRepository
	userEmailRepo               repository.UserEmailRepository
	userCredentialRepo          repository.UserCredentialRepository
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
	userRepo repository.UserRepository,
	userEmailRepo repository.UserEmailRepository,
	userCredentialRepo repository.UserCredentialRepository,
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
		userRepo:                    userRepo,
		userEmailRepo:               userEmailRepo,
		userCredentialRepo:          userCredentialRepo,
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

	email := strings.ToLower(strings.TrimSpace(input.Email))
	emailConfirmation := strings.ToLower(strings.TrimSpace(input.EmailConfirmation))

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

func validateVerifyInput(input VerifyUserRegistrationInput) error {
	fieldErrors := map[string][]app_error.FieldError{}

	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		fieldErrors["display_name"] = append(fieldErrors["display_name"], app_error.FieldError{
			Code: i18n.CodeDisplayNameRequired,
		})
	}

	if input.Password == "" {
		fieldErrors["password"] = append(fieldErrors["password"], app_error.FieldError{
			Code: i18n.CodePasswordRequired,
		})
	}

	if input.PasswordConfirmation == "" {
		fieldErrors["password_confirmation"] = append(fieldErrors["password_confirmation"], app_error.FieldError{
			Code: i18n.CodePasswordConfirmationRequired,
		})
	} else if input.Password != "" && input.Password != input.PasswordConfirmation {
		fieldErrors["password_confirmation"] = append(fieldErrors["password_confirmation"], app_error.FieldError{
			Code: i18n.CodePasswordConfirmationNotMatch,
		})
	}

	if !input.AgreedToTerms {
		fieldErrors["agreed_to_terms"] = append(fieldErrors["agreed_to_terms"], app_error.FieldError{
			Code: i18n.CodeAgreedToTermsRequired,
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

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.EmailConfirmation = strings.ToLower(strings.TrimSpace(input.EmailConfirmation))

	now := s.clock.Now()

	req, err := s.userRegistrationRequestRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if req != nil && req.VerifiedAt != nil {
		return &CreateUserRegistrationOutput{
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: s.config.RegistrationTokenExpiresMinutes(),
		}, nil
	}

	if req != nil && req.LastSentAt != nil {
		resendAvailableAt := req.LastSentAt.Add(
			time.Duration(s.config.RegistrationResendIntervalMinutes()) * time.Minute,
		)
		if now.Before(resendAvailableAt) {
			return &CreateUserRegistrationOutput{
				Code:           i18n.CodeUserRegistrationRequestCreated,
				ExpiresMinutes: s.config.RegistrationTokenExpiresMinutes(),
			}, nil
		}
	}

	plainToken, err := s.tokenGenerator.Generate()
	if err != nil {
		return nil, err
	}
	if plainToken == "" {
		return nil, errors.New("token generator returned empty token")
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

		payload, err := jsonMarshal(map[string]string{
			"email": input.Email,
			"url":   registrationURL,
			"lang":  input.Language,
		})
		if err != nil {
			return err
		}

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
		Code:           i18n.CodeUserRegistrationRequestCreated,
		ExpiresMinutes: s.config.RegistrationTokenExpiresMinutes(),
	}, nil
}

func (s *userRegistrationService) Verify(
	ctx context.Context,
	input VerifyUserRegistrationInput,
) (*VerifyUserRegistrationOutput, error) {

	if input.Token == "" {
		return nil, app_error.NewBadRequest(i18n.CodeInvalidRegistrationToken)
	}

	if err := validateVerifyInput(input); err != nil {
		return nil, err
	}

	input.DisplayName = strings.TrimSpace(input.DisplayName)

	tokenHash, err := s.tokenHasher.Hash(input.Token)
	if err != nil {
		return nil, app_error.WrapInternal(err)
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, app_error.WrapInternal(err)
	}

	now := s.clock.Now()

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		req, err := s.userRegistrationRequestRepo.FindByTokenHashForUpdate(txCtx, tokenHash)
		if err != nil {
			return err
		}

		if req == nil {
			return app_error.NewBadRequest(i18n.CodeInvalidRegistrationToken)
		}

		if req.VerifiedAt != nil {
			return app_error.NewConflict(i18n.CodeUsedRegistrationToken)
		}

		if now.After(req.ExpiresAt) {
			return app_error.NewBadRequest(i18n.CodeExpiredRegistrationToken)
		}

		existing, err := s.userEmailRepo.FindByEmail(txCtx, req.Email)
		if err != nil {
			return err
		}
		if existing != nil {
			return app_error.NewConflict(i18n.CodeUserAlreadyRegistered)
		}

		userID, err := s.uuidGenerator.NewV7()
		if err != nil {
			return err
		}

		if err := s.userRepo.Create(txCtx, &model.User{
			ID:          userID,
			DisplayName: input.DisplayName,
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		}); err != nil {
			return err
		}

		emailID, err := s.uuidGenerator.NewV7()
		if err != nil {
			return err
		}

		if err := s.userEmailRepo.Create(txCtx, &model.UserEmail{
			ID:         emailID,
			UserID:     userID,
			Email:      req.Email,
			IsPrimary:  true,
			VerifiedAt: &now,
			CreatedAt:  now,
			UpdatedAt:  now,
		}); err != nil {
			return err
		}

		if err := s.userCredentialRepo.Create(txCtx, &model.UserCredential{
			UserID:            userID,
			PasswordHash:      passwordHash,
			PasswordChangedAt: now,
			CreatedAt:         now,
		}); err != nil {
			return err
		}

		req.VerifiedAt = &now
		return s.userRegistrationRequestRepo.UpdateVerifiedAt(txCtx, req)
	})

	if err != nil {
		return nil, err
	}

	return &VerifyUserRegistrationOutput{
		Code: i18n.CodeUserRegistrationVerified,
	}, nil
}
