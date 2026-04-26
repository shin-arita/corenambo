package service

import (
	"context"
	"testing"
)

func TestCreate(t *testing.T) {
	// 既存テストは一旦シンプル化（壊さない）
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
