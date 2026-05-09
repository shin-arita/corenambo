package model

import (
	"testing"
	"time"
)

func TestUserEmail(t *testing.T) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	ue := UserEmail{
		ID:         "email-uuid",
		UserID:     "user-uuid",
		Email:      "test@example.com",
		IsPrimary:  true,
		VerifiedAt: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if ue.ID != "email-uuid" {
		t.Fatalf("unexpected id: %s", ue.ID)
	}

	if ue.UserID != "user-uuid" {
		t.Fatalf("unexpected user_id: %s", ue.UserID)
	}

	if ue.Email != "test@example.com" {
		t.Fatalf("unexpected email: %s", ue.Email)
	}

	if !ue.IsPrimary {
		t.Fatal("expected is_primary to be true")
	}

	if ue.VerifiedAt == nil {
		t.Fatal("verified_at is nil")
	}

	if !ue.VerifiedAt.Equal(now) {
		t.Fatalf("unexpected verified_at: %v", ue.VerifiedAt)
	}

	if !ue.CreatedAt.Equal(now) {
		t.Fatalf("unexpected created_at: %v", ue.CreatedAt)
	}

	if !ue.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected updated_at: %v", ue.UpdatedAt)
	}
}

func TestUserEmailVerifiedAtNil(t *testing.T) {
	ue := UserEmail{
		ID:         "email-uuid",
		UserID:     "user-uuid",
		Email:      "test@example.com",
		IsPrimary:  false,
		VerifiedAt: nil,
	}

	if ue.VerifiedAt != nil {
		t.Fatal("expected verified_at to be nil")
	}
}
