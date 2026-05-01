package mail

import (
	"context"
	"errors"
	"net/smtp"
	"strings"
	"testing"

	"app-api/internal/i18n"
)

func TestNewSMTPMailer(t *testing.T) {
	mailer := NewSMTPMailer("mail", "1025", "noreply@example.com", "", "", false)

	smtpMailer, ok := mailer.(*SMTPMailer)
	if !ok {
		t.Fatal("mailer is not SMTPMailer")
	}

	if smtpMailer.Host != "mail" {
		t.Fatalf("unexpected host: %s", smtpMailer.Host)
	}

	if smtpMailer.Port != "1025" {
		t.Fatalf("unexpected port: %s", smtpMailer.Port)
	}

	if smtpMailer.From != "noreply@example.com" {
		t.Fatalf("unexpected from: %s", smtpMailer.From)
	}
}

func TestSMTPMailerSendUserRegistrationMail(t *testing.T) {
	var capturedAddr string
	var capturedFrom string
	var capturedTo []string
	var capturedMsg string

	mailer := &SMTPMailer{
		Host: "mail",
		Port: "1025",
		From: "noreply@example.com",
		sendMail: func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
			capturedAddr = addr
			capturedFrom = from
			capturedTo = to
			capturedMsg = string(msg)
			return nil
		},
		tl: i18n.NewTranslator(),
	}

	err := mailer.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:             "test@example.com",
		URL:            "http://example.com/verify?token=abc",
		Lang:           "ja",
		ExpiresMinutes: 60,
	})
	if err != nil {
		t.Fatal(err)
	}

	if capturedAddr != "mail:1025" {
		t.Fatalf("unexpected addr: %s", capturedAddr)
	}

	if !strings.Contains(capturedMsg, "From: noreply@example.com") {
		t.Fatal("from header not found")
	}

	if !strings.Contains(capturedMsg, "To: test@example.com") {
		t.Fatal("to header not found")
	}

	if capturedFrom != "noreply@example.com" {
		t.Fatalf("unexpected from: %s", capturedFrom)
	}

	if len(capturedTo) != 1 || capturedTo[0] != "test@example.com" {
		t.Fatalf("unexpected to: %v", capturedTo)
	}

	if !strings.Contains(capturedMsg, "http://example.com/verify?token=abc") {
		t.Fatalf("url not found in message: %s", capturedMsg)
	}

	if !strings.Contains(capturedMsg, "Subject:") {
		t.Fatal("subject not found")
	}
}

func TestSMTPMailerSendUserRegistrationMailWithAuth(t *testing.T) {
	var capturedAuth smtp.Auth

	mailer := &SMTPMailer{
		Host: "mail",
		Port: "1025",
		From: "noreply@example.com",
		User: "user@example.com",
		Pass: "secret",
		sendMail: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			capturedAuth = a
			return nil
		},
		tl: i18n.NewTranslator(),
	}

	err := mailer.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:             "test@example.com",
		URL:            "http://example.com/verify?token=abc",
		Lang:           "ja",
		ExpiresMinutes: 60,
	})
	if err != nil {
		t.Fatal(err)
	}

	if capturedAuth == nil {
		t.Fatal("auth should not be nil when User is set")
	}
}

func TestSMTPMailerSendUserRegistrationMailWithTLS(t *testing.T) {
	var capturedAddr string
	var capturedAuth smtp.Auth

	mailer := &SMTPMailer{
		Host:   "mail",
		Port:   "465",
		From:   "noreply@example.com",
		User:   "user@example.com",
		Pass:   "secret",
		UseTLS: true,
		sendTLSMail: func(addr string, auth smtp.Auth, from string, to string, message []byte) error {
			capturedAddr = addr
			capturedAuth = auth
			return nil
		},
		tl: i18n.NewTranslator(),
	}

	err := mailer.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:             "test@example.com",
		URL:            "http://example.com/verify?token=abc",
		Lang:           "ja",
		ExpiresMinutes: 60,
	})
	if err != nil {
		t.Fatal(err)
	}

	if capturedAddr != "mail:465" {
		t.Fatalf("unexpected addr: %s", capturedAddr)
	}

	if capturedAuth == nil {
		t.Fatal("auth should not be nil when User is set")
	}
}

func TestSMTPMailerSendUserRegistrationMailWithTLSNoAuth(t *testing.T) {
	var capturedAuth smtp.Auth

	mailer := &SMTPMailer{
		Host:   "mail",
		Port:   "465",
		From:   "noreply@example.com",
		UseTLS: true,
		sendTLSMail: func(addr string, auth smtp.Auth, from string, to string, message []byte) error {
			capturedAuth = auth
			return nil
		},
		tl: i18n.NewTranslator(),
	}

	err := mailer.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:             "test@example.com",
		URL:            "http://example.com/verify?token=abc",
		Lang:           "ja",
		ExpiresMinutes: 60,
	})
	if err != nil {
		t.Fatal(err)
	}

	if capturedAuth != nil {
		t.Fatal("auth should be nil when User is not set")
	}
}

func TestSMTPMailerSendUserRegistrationMailSendError(t *testing.T) {
	expectedErr := errors.New("send failed")

	mailer := &SMTPMailer{
		Host: "mail",
		Port: "1025",
		From: "noreply@example.com",
		sendMail: func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
			return expectedErr
		},
		tl: i18n.NewTranslator(),
	}

	err := mailer.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:             "test@example.com",
		URL:            "http://example.com/verify?token=abc",
		Lang:           "ja",
		ExpiresMinutes: 60,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

type brokenTranslator struct{}

func (t *brokenTranslator) Translate(lang string, code string) string {
	return "{{"
}

func TestSMTPMailerSendUserRegistrationMailTemplateParseError(t *testing.T) {
	mailer := &SMTPMailer{
		Host:     "mail",
		Port:     "1025",
		From:     "noreply@example.com",
		sendMail: smtp.SendMail,
		tl:       &brokenTranslator{},
	}

	err := mailer.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:             "test@example.com",
		URL:            "http://example.com/verify?token=abc",
		Lang:           "ja",
		ExpiresMinutes: 60,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

type executeErrorTranslator struct{}

func (t *executeErrorTranslator) Translate(lang string, code string) string {
	return "{{.Missing.Field}}"
}

func TestSMTPMailerSendUserRegistrationMailTemplateExecuteError(t *testing.T) {
	mailer := &SMTPMailer{
		Host:     "mail",
		Port:     "1025",
		From:     "noreply@example.com",
		sendMail: smtp.SendMail,
		tl:       &executeErrorTranslator{},
	}

	err := mailer.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:             "test@example.com",
		URL:            "http://example.com/verify?token=abc",
		Lang:           "ja",
		ExpiresMinutes: 60,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSMTPMailerSendUserAlreadyRegisteredMail(t *testing.T) {
	var capturedMsg string

	mailer := &SMTPMailer{
		Host: "mail",
		Port: "1025",
		From: "noreply@example.com",
		sendMail: func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
			capturedMsg = string(msg)
			return nil
		},
		tl: i18n.NewTranslator(),
	}

	err := mailer.SendUserAlreadyRegisteredMail(context.Background(), "test@example.com", "ja")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(capturedMsg, "既に登録されています") {
		t.Fatalf("unexpected message: %s", capturedMsg)
	}
}

func TestSMTPMailerSendUserAlreadyRegisteredMailSendError(t *testing.T) {
	expectedErr := errors.New("send failed")

	mailer := &SMTPMailer{
		Host: "mail",
		Port: "1025",
		From: "noreply@example.com",
		sendMail: func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
			return expectedErr
		},
		tl: i18n.NewTranslator(),
	}

	err := mailer.SendUserAlreadyRegisteredMail(context.Background(), "test@example.com", "ja")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}
