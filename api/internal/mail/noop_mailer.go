package mail

import "context"

type NoopMailer struct{}

func (m *NoopMailer) SendUserRegistrationMail(ctx context.Context, mail UserRegistrationMail) error {
	return nil
}
