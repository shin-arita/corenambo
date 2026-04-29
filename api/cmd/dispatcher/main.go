package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"app-api/internal/mail"
	"app-api/internal/repository"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://app_user:password@db:5432/app_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewMailOutboxRepository(db)
	
	var mailer mail.Mailer = &mail.NoopMailer{}

	ctx := context.Background()

	pending, err := repo.FetchPending(ctx, 100)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range pending {

		msg := mail.UserRegistrationMail{
			To: p.ToEmail,
		}

		err := mailer.SendUserRegistrationMail(ctx, msg)
		if err != nil {
			_ = repo.MarkFailed(ctx, p.ID, err.Error(), time.Now().Add(5*time.Minute))
			continue
		}

		_ = repo.MarkSent(ctx, p.ID, time.Now())
	}
}
