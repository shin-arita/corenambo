package model

import "time"

type CreateUserRegistrationRequestInput struct {
	Email             string `json:"email"`
	EmailConfirmation string `json:"email_confirmation"`
}

type UserRegistrationRequest struct {
	ID         string
	Email      string
	TokenHash  string
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	LastSentAt *time.Time
	CreatedAt  time.Time
}

type UserRegistrationRequestServiceResult struct {
	Code string
}
