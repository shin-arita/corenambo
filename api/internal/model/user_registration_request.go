package model

import "time"

type UserRegistrationRequest struct {
	ID         string
	Email      string
	TokenHash  string
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	LastSentAt *time.Time
	CreatedAt  time.Time
}
