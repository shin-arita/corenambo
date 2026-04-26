package model

import "time"

type MailOutbox struct {
	ID            string
	MailType      string
	ToEmail       string
	Payload       string
	Status        string
	RetryCount    int
	NextAttemptAt time.Time
	SentAt        *time.Time
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
