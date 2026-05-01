package mail

import (
	"bytes"
	"context"
	"net/smtp"
	"text/template"

	"app-api/internal/i18n"
)

type SMTPMailer struct {
	Host        string
	Port        string
	From        string
	User        string
	Pass        string
	UseTLS      bool
	sendMail    func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	sendTLSMail func(addr string, auth smtp.Auth, from string, to string, message []byte) error
	tl          i18n.Translator
}

func NewSMTPMailer(host, port, from, user, pass string, useTLS bool) Mailer {
	return &SMTPMailer{
		Host:        host,
		Port:        port,
		From:        from,
		User:        user,
		Pass:        pass,
		UseTLS:      useTLS,
		sendMail:    smtp.SendMail,
		sendTLSMail: defaultSendWithTLS,
		tl:          i18n.NewTranslator(),
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

	addr := m.Host + ":" + m.Port

	// 認証情報が設定されている場合のみ認証を使用（開発/本番切替）
	var auth smtp.Auth
	if m.User != "" {
		auth = smtp.PlainAuth("", m.User, m.Pass, m.Host)
	}

	if m.UseTLS {
		return m.sendTLSMail(addr, auth, m.From, to, message)
	}

	// 開発環境（Mailpit等）: 認証なし・TLSなし
	return m.sendMail(addr, auth, m.From, []string{to}, message)
}
