package model

import (
	"testing"
	"time"
)

func TestUserRegistrationRequest(t *testing.T) {
	verifiedAt := time.Date(2026, 4, 24, 18, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2026, 4, 24, 19, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 24, 18, 0, 0, 0, time.UTC)

	req := UserRegistrationRequest{
		ID:         "uuid-v7",
		Email:      "test@example.com",
		TokenHash:  "hashed-token",
		ExpiresAt:  expiresAt,
		VerifiedAt: &verifiedAt,
		CreatedAt:  createdAt,
	}

	if req.ID != "uuid-v7" {
		t.Fatalf("unexpected id: %s", req.ID)
	}

	if req.Email != "test@example.com" {
		t.Fatalf("unexpected email: %s", req.Email)
	}

	if req.TokenHash != "hashed-token" {
		t.Fatalf("unexpected token hash: %s", req.TokenHash)
	}

	if !req.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected expires at: %v", req.ExpiresAt)
	}

	if req.VerifiedAt == nil {
		t.Fatal("verified at is nil")
	}

	if !req.VerifiedAt.Equal(verifiedAt) {
		t.Fatalf("unexpected verified at: %v", req.VerifiedAt)
	}

	if !req.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected created at: %v", req.CreatedAt)
	}
}
