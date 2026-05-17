package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/unicode/norm"

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
	CheckToken(ctx context.Context, input CheckTokenInput) (*CheckTokenOutput, error)
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

type CheckTokenInput struct {
	Token string
}

type CheckTokenOutput struct {
	Code string
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
	} else if addr, err := mail.ParseAddress(email); err != nil || addr.Name != "" || addr.Address != email {
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

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

// reservedDisplayNames は登録禁止のユーザ名リスト（日本語・英語・中国語）。
// NFC+trim 済みの値に対して strings.ToLower で大小文字を正規化して完全一致判定する。
var reservedDisplayNames = []string{
	// 日本語
	"管理者", "運営", "公式", "サポート", "システム",
	// 英語
	"admin", "administrator", "official", "support", "system", "root",
	// 中国語
	"管理员", "官方", "客服", "系统",
}

func isReservedDisplayName(name string) bool {
	lower := strings.ToLower(name)
	for _, r := range reservedDisplayNames {
		if lower == strings.ToLower(r) {
			return true
		}
	}
	return false
}

// isControlChar は制御文字・改行・行区切り文字を検出する。
func isControlChar(r rune) bool {
	// unicode.IsControl は C0 (U+0000-U+001F) / DEL (U+007F) / C1 (U+0080-U+009F) を対象とする。
	// U+2028 LINE SEPARATOR / U+2029 PARAGRAPH SEPARATOR は改行相当のため追加で禁止する。
	return unicode.IsControl(r) || r == 0x2028 || r == 0x2029
}

// isDisallowedZeroWidth は ZWJ (U+200D) を除くゼロ幅文字を検出する。
func isDisallowedZeroWidth(r rune) bool {
	switch r {
	case 0x200B, // ZERO WIDTH SPACE
		0x200C, // ZERO WIDTH NON-JOINER
		0xFEFF, // ZERO WIDTH NO-BREAK SPACE / BOM
		0x2060, // WORD JOINER
		0x2061, // FUNCTION APPLICATION
		0x2062, // INVISIBLE TIMES
		0x2063, // INVISIBLE SEPARATOR
		0x2064, // INVISIBLE PLUS
		0x034F, // COMBINING GRAPHEME JOINER
		0x00AD: // SOFT HYPHEN
		return true
	}
	return false
}

// validateDisplayName は NFC正規化・trim 済みの display_name に対して文字仕様を検証する。
// 3〜30文字、制御文字禁止、ZWJ以外のゼロ幅文字禁止。
func validateDisplayName(name string) *app_error.FieldError {
	runes := []rune(name)
	if len(runes) < 3 {
		fe := app_error.FieldError{Code: i18n.CodeDisplayNameTooShort}
		return &fe
	}
	if len(runes) > 30 {
		fe := app_error.FieldError{Code: i18n.CodeDisplayNameTooLong}
		return &fe
	}
	for _, r := range runes {
		if isControlChar(r) {
			fe := app_error.FieldError{Code: i18n.CodeDisplayNameControlChar}
			return &fe
		}
		if isDisallowedZeroWidth(r) {
			fe := app_error.FieldError{Code: i18n.CodeDisplayNameZeroWidth}
			return &fe
		}
	}
	if isReservedDisplayName(name) {
		fe := app_error.FieldError{Code: i18n.CodeDisplayNameReserved}
		return &fe
	}
	return nil
}

func validateVerifyInput(input VerifyUserRegistrationInput) error {
	fieldErrors := map[string][]app_error.FieldError{}

	displayName := strings.TrimSpace(norm.NFC.String(input.DisplayName))
	if displayName == "" {
		fieldErrors["display_name"] = append(fieldErrors["display_name"], app_error.FieldError{
			Code: i18n.CodeDisplayNameRequired,
		})
	} else if fe := validateDisplayName(displayName); fe != nil {
		fieldErrors["display_name"] = append(fieldErrors["display_name"], *fe)
	}

	if input.Password == "" {
		fieldErrors["password"] = append(fieldErrors["password"], app_error.FieldError{
			Code: i18n.CodePasswordRequired,
		})
	} else if len(input.Password) > 72 {
		fieldErrors["password"] = append(fieldErrors["password"], app_error.FieldError{
			Code: i18n.CodePasswordTooLong,
		})
	} else if !isStrongPassword(input.Password) {
		fieldErrors["password"] = append(fieldErrors["password"], app_error.FieldError{
			Code: i18n.CodePasswordTooWeak,
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

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// FOR UPDATE でロック取得し、concurrent再送による二重メール送信を防ぐ
		req, err := s.userRegistrationRequestRepo.FindByEmailForUpdate(txCtx, input.Email)
		if err != nil {
			return err
		}

		// 認証済みの場合はそのまま成功を返す
		if req != nil && req.VerifiedAt != nil {
			return nil
		}

		// 再送インターバル内の場合はそのまま成功を返す
		if req != nil && req.LastSentAt != nil {
			resendAvailableAt := req.LastSentAt.Add(
				time.Duration(s.config.RegistrationResendIntervalMinutes()) * time.Minute,
			)
			if now.Before(resendAvailableAt) {
				return nil
			}
		}

		// 送信が必要と確定してからトークン材料を生成（純粋な計算処理）
		plainToken, err := s.tokenGenerator.Generate()
		if err != nil {
			return err
		}
		if plainToken == "" {
			return errors.New("token generator returned empty token")
		}

		tokenHash, err := s.tokenHasher.Hash(plainToken)
		if err != nil {
			return err
		}

		expiresAt := now.Add(
			time.Duration(s.config.RegistrationTokenExpiresMinutes()) * time.Minute,
		)

		registrationURL := s.registrationURLBuilder.Build(plainToken)

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
				// concurrent INSERTによるuniqueviolationは成功として扱う
				if errors.Is(err, repository.ErrDuplicateEmail) {
					return nil
				}
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

	input.DisplayName = strings.TrimSpace(norm.NFC.String(input.DisplayName))

	tokenHash, err := s.tokenHasher.Hash(input.Token)
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

		// token有効確認後にbcryptを実行（無効token大量投入によるCPU負荷を防ぐ）
		passwordHash, err := hashPassword(input.Password)
		if err != nil {
			return app_error.WrapInternal(err)
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

func (s *userRegistrationService) CheckToken(
	ctx context.Context,
	input CheckTokenInput,
) (*CheckTokenOutput, error) {

	if input.Token == "" {
		return nil, app_error.NewBadRequest(i18n.CodeInvalidRegistrationToken)
	}

	tokenHash, err := s.tokenHasher.Hash(input.Token)
	if err != nil {
		return nil, app_error.WrapInternal(err)
	}

	now := s.clock.Now()

	req, err := s.userRegistrationRequestRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if req == nil {
		return nil, app_error.NewBadRequest(i18n.CodeInvalidRegistrationToken)
	}

	if req.VerifiedAt != nil {
		return nil, app_error.NewConflict(i18n.CodeUsedRegistrationToken)
	}

	if now.After(req.ExpiresAt) {
		return nil, app_error.NewBadRequest(i18n.CodeExpiredRegistrationToken)
	}

	existing, err := s.userEmailRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, app_error.NewConflict(i18n.CodeUserAlreadyRegistered)
	}

	return &CheckTokenOutput{Code: i18n.CodeRegistrationTokenValid}, nil
}
