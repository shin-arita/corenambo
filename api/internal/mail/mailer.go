package mail

import "context"

type Mailer interface {
	SendUserRegistrationMail(ctx context.Context, mail UserRegistrationMail) error
	SendUserAlreadyRegisteredMail(ctx context.Context, to string, lang string) error
}

// NonRetryableMailError はリトライしても解消しないメールエラーを表す。
// payload 不備・テンプレートエラーなど送信前に確定できる失敗に使用する。
// worker は errors.As でこの型を検出し MarkFailed を呼ぶ。
type NonRetryableMailError struct {
	Msg string
}

func (e *NonRetryableMailError) Error() string { return e.Msg }
