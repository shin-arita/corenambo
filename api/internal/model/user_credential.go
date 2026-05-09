package model

import "time"

type UserCredential struct {
	UserID            string
	PasswordHash      string
	PasswordChangedAt time.Time
	CreatedAt         time.Time
}
