package model

import "time"

type User struct {
	ID          string
	DisplayName string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
