package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"app-api/internal/config"
	"app-api/internal/mail"
	"app-api/internal/repository"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.NewWorkerConfig()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewMailOutboxRepository(db)
	mailer := mail.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPUseTLS)

	stuckTimeout := time.Duration(cfg.WorkerStuckTimeoutMinutes()) * time.Minute

	ctx := context.Background()
	for {
		run(ctx, repo, mailer, cfg, stuckTimeout)
		time.Sleep(1 * time.Second)
	}
}

func run(
	ctx context.Context,
	repo repository.MailOutboxRepository,
	mailer mail.Mailer,
	cfg config.WorkerConfig,
	stuckTimeout time.Duration,
) {
	stuckBefore := time.Now().Add(-stuckTimeout)
	if err := repo.ResetStuckProcessing(ctx, stuckBefore); err != nil {
		log.Println("reset stuck processing error:", err)
	}

	pending, err := repo.FetchPending(ctx, 100)
	if err != nil {
		log.Println("fetch error:", err)
		return
	}

	for _, p := range pending {
		if p.RetryCount >= cfg.WorkerMaxRetryCount() {
			log.Printf("[outbox] max retry reached id=%s retry_count=%d", p.ID, p.RetryCount)
			_ = repo.MarkFailed(ctx, p.ID, "max retry count reached", time.Now().Add(24*time.Hour))
			continue
		}

		if err := repo.MarkProcessing(ctx, p.ID); err != nil {
			continue
		}

		log.Printf("[outbox] processing id=%s", p.ID)

		var payload map[string]string
		if err := json.Unmarshal([]byte(p.Payload), &payload); err != nil {
			_ = repo.MarkFailed(ctx, p.ID, "invalid payload", time.Now().Add(5*time.Minute))
			continue
		}

		err := mailer.SendUserRegistrationMail(ctx, mail.UserRegistrationMail{
			To:             p.ToEmail,
			URL:            payload["url"],
			Lang:           payload["lang"],
			ExpiresMinutes: cfg.WorkerRegistrationExpiresMinutes(),
		})
		if err != nil {
			log.Printf("[outbox] failed id=%s err=%s", p.ID, err.Error())
			_ = repo.MarkFailed(ctx, p.ID, err.Error(), time.Now().Add(5*time.Minute))
			continue
		}

		if err := repo.MarkSent(ctx, p.ID, time.Now()); err != nil {
			log.Printf("[outbox] mark sent failed id=%s", p.ID)
		}
		log.Printf("[outbox] sent id=%s", p.ID)
	}
}
