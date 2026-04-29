package service

import (
	"context"
	"testing"
)

func TestCreate(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate_EmailEmpty(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "",
		EmailConfirmation: "",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_EmailFormatInvalid(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "not-an-email",
		EmailConfirmation: "not-an-email",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_EmailConfirmationNotMatch(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "other@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_EmailConfirmationEmpty(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_AlreadyVerified(t *testing.T) {
	svc := newTestUserRegistrationServiceWithVerifiedUser()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate_ResendNotAvailable(t *testing.T) {
	svc := newTestUserRegistrationServiceWithRecentlySentUser()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate_UpdateToken(t *testing.T) {
	svc := newTestUserRegistrationServiceWithExpiredUser()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate_DBError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithDBError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_TokenGenError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithTokenGenError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_TokenHashError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithTokenHashError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_FirstUUIDError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithFirstUUIDError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_SecondUUIDError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithSecondUUIDError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_CreateUserError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithCreateUserError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_UpdateTokenError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithUpdateTokenError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerify(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Verify(context.Background(), VerifyUserRegistrationInput{
		Token: "token",
	})
	if err != nil {
		t.Fatal(err)
	}
}
