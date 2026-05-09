package model

import "time"

type UserEmail struct {
	ID         string
	UserID     string
	Email      string
	IsPrimary  bool
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
