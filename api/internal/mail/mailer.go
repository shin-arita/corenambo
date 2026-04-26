package mail

import "context"

type Mailer interface {
	SendUserRegistrationMail(ctx context.Context, mail UserRegistrationMail) error
	SendUserAlreadyRegisteredMail(ctx context.Context, to string, lang string) error
}
