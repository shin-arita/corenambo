package registrationurl

import "fmt"

type Builder interface {
	Build(token string) string
}

type StaticBuilder struct {
	FrontendBaseURL string
}

func NewStaticBuilder(frontendBaseURL string) Builder {
	return &StaticBuilder{
		FrontendBaseURL: frontendBaseURL,
	}
}

func (b *StaticBuilder) Build(token string) string {
	return fmt.Sprintf("%s/user-registration/verify?token=%s", b.FrontendBaseURL, token)
}
