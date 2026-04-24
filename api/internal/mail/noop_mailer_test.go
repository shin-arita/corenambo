package mail

import (
	"context"
	"testing"
)

func TestNoopMailerSendUserRegistrationMail(t *testing.T) {
	m := &NoopMailer{}

	err := m.SendUserRegistrationMail(context.Background(), UserRegistrationMail{
		To:   "test@example.com",
		URL:  "http://example.com/register?token=test-token",
		Lang: "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
}
