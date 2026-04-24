package mail

import "context"

type Mailer interface {
	SendUserRegistrationMail(ctx context.Context, mail UserRegistrationMail) error
}
