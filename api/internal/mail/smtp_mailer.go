package mail

import (
	"bytes"
	"context"
	"net/smtp"
	"text/template"

	"app-api/internal/i18n"
)

type SMTPMailer struct {
	Host     string
	Port     string
	From     string
	sendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	tl       i18n.Translator
}

func NewSMTPMailer(host string, port string, from string) Mailer {
	return &SMTPMailer{
		Host:     host,
		Port:     port,
		From:     from,
		sendMail: smtp.SendMail,
		tl:       i18n.NewTranslator(),
	}
}

func (m *SMTPMailer) SendUserRegistrationMail(ctx context.Context, mailData UserRegistrationMail) error {
	subjectText := m.tl.Translate(mailData.Lang, i18n.CodeMailUserRegistrationSubject)
	bodyTemplate := m.tl.Translate(mailData.Lang, i18n.CodeMailUserRegistrationBody)

	tmpl, err := template.New("mail").Option("missingkey=error").Parse(bodyTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"URL":            mailData.URL,
		"ExpiresMinutes": mailData.ExpiresMinutes,
	})
	if err != nil {
		return err
	}

	return m.send(mailData.To, subjectText, buf.String())
}

func (m *SMTPMailer) SendUserAlreadyRegisteredMail(ctx context.Context, to string, lang string) error {
	subjectText := m.tl.Translate(lang, i18n.CodeMailUserAlreadyRegisteredSubject)
	bodyText := m.tl.Translate(lang, i18n.CodeMailUserAlreadyRegisteredBody)

	return m.send(to, subjectText, bodyText)
}

func (m *SMTPMailer) send(to string, subjectText string, body string) error {
	from := "From: " + m.From + "\r\n"
	toHeader := "To: " + to + "\r\n"
	subject := "Subject: " + subjectText + "\r\n"
	contentType := "Content-Type: text/plain; charset=UTF-8\r\n"

	message := []byte(
		from +
			toHeader +
			subject +
			contentType +
			"\r\n" +
			body,
	)

	return m.sendMail(
		m.Host+":"+m.Port,
		nil,
		m.From,
		[]string{to},
		message,
	)
}
