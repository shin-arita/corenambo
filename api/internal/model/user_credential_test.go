package model

import (
	"testing"
	"time"
)

func TestUserCredential(t *testing.T) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	uc := UserCredential{
		UserID:            "user-uuid",
		PasswordHash:      "$2a$10$hashedpassword",
		PasswordChangedAt: now,
		CreatedAt:         now,
	}

	if uc.UserID != "user-uuid" {
		t.Fatalf("unexpected user_id: %s", uc.UserID)
	}

	if uc.PasswordHash != "$2a$10$hashedpassword" {
		t.Fatalf("unexpected password_hash: %s", uc.PasswordHash)
	}

	if !uc.PasswordChangedAt.Equal(now) {
		t.Fatalf("unexpected password_changed_at: %v", uc.PasswordChangedAt)
	}

	if !uc.CreatedAt.Equal(now) {
		t.Fatalf("unexpected created_at: %v", uc.CreatedAt)
	}
}
