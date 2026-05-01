package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

// defaultSendWithTLS はSMTPS（ポート465）で直接TLS接続してメールを送信する
// 実際のTLS接続が必要なためユニットテストはsendTLSMailフィールドでモックすること
func defaultSendWithTLS(addr string, auth smtp.Auth, from string, to string, message []byte) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("smtp split host port: %w", err)
	}

	tlsCfg := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write(message); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}

	return client.Quit()
}
