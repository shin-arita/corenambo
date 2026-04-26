package mail

import "context"

type NoopMailer struct{}

func (m *NoopMailer) SendUserRegistrationMail(ctx context.Context, mail UserRegistrationMail) error {
	return nil
}

func (m *NoopMailer) SendUserAlreadyRegisteredMail(ctx context.Context, to string, lang string) error {
	return nil
}
