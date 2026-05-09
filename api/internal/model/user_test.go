package model

import (
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	u := User{
		ID:          "uuid-v7",
		DisplayName: "testuser",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if u.ID != "uuid-v7" {
		t.Fatalf("unexpected id: %s", u.ID)
	}

	if u.DisplayName != "testuser" {
		t.Fatalf("unexpected display_name: %s", u.DisplayName)
	}

	if u.Status != "active" {
		t.Fatalf("unexpected status: %s", u.Status)
	}

	if !u.CreatedAt.Equal(now) {
		t.Fatalf("unexpected created_at: %v", u.CreatedAt)
	}

	if !u.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected updated_at: %v", u.UpdatedAt)
	}
}
