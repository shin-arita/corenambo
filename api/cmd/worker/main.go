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
	mailer := &mail.NoopMailer{}

	ctx := context.Background()

	for {
		run(ctx, repo, mailer)
		time.Sleep(3 * time.Second)
	}
}

func run(
	ctx context.Context,
	repo repository.MailOutboxRepository, // ★ここ修正（ポインタ禁止）
	mailer mail.Mailer,
) {
	pending, err := repo.FetchPending(ctx, 100)
	if err != nil {
		log.Println("fetch error:", err)
		return
	}

	for _, p := range pending {

		err := mailer.SendUserRegistrationMail(ctx, mail.UserRegistrationMail{
			To: p.ToEmail,
		})

		if err != nil {
			_ = repo.MarkFailed(ctx, p.ID, err.Error(), time.Now().Add(5*time.Minute))
			continue
		}

		_ = repo.MarkSent(ctx, p.ID, time.Now())
	}
}
