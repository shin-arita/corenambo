package mail

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"testing"
)

func serveFakeSMTP(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	r := bufio.NewReader(conn)
	write := func(s string) { _, _ = fmt.Fprintf(conn, "%s\r\n", s) }

	write("220 localhost ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		parts := strings.Fields(strings.TrimRight(line, "\r\n"))
		if len(parts) == 0 {
			continue
		}
		switch strings.ToUpper(parts[0]) {
		case "EHLO", "HELO":
			write("250 localhost")
		case "AUTH":
			write("235 2.7.0 OK")
		case "MAIL":
			write("250 OK")
		case "RCPT":
			write("250 OK")
		case "DATA":
			write("354 Start mail input")
			for {
				l, err := r.ReadString('\n')
				if err != nil || strings.TrimRight(l, "\r\n") == "." {
					break
				}
			}
			write("250 OK")
		case "QUIT":
			write("221 Bye")
			return
		}
	}
}

func withFakeConn(handler func(net.Conn)) net.Conn {
	server, client := net.Pipe()
	go handler(server)
	return client
}

func TestDefaultDialTLS_Error(t *testing.T) {
	_, err := defaultDialTLS("tcp", "localhost:0", &tls.Config{})
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestDefaultSendWithTLS_InvalidAddr(t *testing.T) {
	err := defaultSendWithTLS("invalid-no-port", nil, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected error for invalid addr")
	}
	if !strings.Contains(err.Error(), "smtp split host port") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultSendWithTLS_DialError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return nil, errors.New("mock dial error")
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected dial error")
	}
	if !strings.Contains(err.Error(), "smtp tls dial") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultSendWithTLS_NewClientError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		server, client := net.Pipe()
		_ = server.Close()
		return client, nil
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected smtp.NewClient error")
	}
	if !strings.Contains(err.Error(), "smtp new client") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultSendWithTLS_MailError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(func(conn net.Conn) {
			defer func() { _ = conn.Close() }()
			r := bufio.NewReader(conn)
			_, _ = fmt.Fprintf(conn, "220 localhost ESMTP\r\n")
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					return
				}
				cmd := strings.ToUpper(strings.Fields(strings.TrimRight(line, "\r\n"))[0])
				switch cmd {
				case "EHLO", "HELO":
					_, _ = fmt.Fprintf(conn, "250 localhost\r\n")
				case "MAIL":
					_, _ = fmt.Fprintf(conn, "550 Rejected\r\n")
					return
				}
			}
		}), nil
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected MAIL FROM error")
	}
}

func TestDefaultSendWithTLS_RcptError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(func(conn net.Conn) {
			defer func() { _ = conn.Close() }()
			r := bufio.NewReader(conn)
			_, _ = fmt.Fprintf(conn, "220 localhost ESMTP\r\n")
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					return
				}
				cmd := strings.ToUpper(strings.Fields(strings.TrimRight(line, "\r\n"))[0])
				switch cmd {
				case "EHLO", "HELO":
					_, _ = fmt.Fprintf(conn, "250 localhost\r\n")
				case "MAIL":
					_, _ = fmt.Fprintf(conn, "250 OK\r\n")
				case "RCPT":
					_, _ = fmt.Fprintf(conn, "550 Rejected\r\n")
					return
				}
			}
		}), nil
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected RCPT TO error")
	}
}

func TestDefaultSendWithTLS_DataError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(func(conn net.Conn) {
			defer func() { _ = conn.Close() }()
			r := bufio.NewReader(conn)
			_, _ = fmt.Fprintf(conn, "220 localhost ESMTP\r\n")
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					return
				}
				cmd := strings.ToUpper(strings.Fields(strings.TrimRight(line, "\r\n"))[0])
				switch cmd {
				case "EHLO", "HELO":
					_, _ = fmt.Fprintf(conn, "250 localhost\r\n")
				case "MAIL":
					_, _ = fmt.Fprintf(conn, "250 OK\r\n")
				case "RCPT":
					_, _ = fmt.Fprintf(conn, "250 OK\r\n")
				case "DATA":
					_, _ = fmt.Fprintf(conn, "452 Insufficient storage\r\n")
					return
				}
			}
		}), nil
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected DATA error")
	}
}

func TestDefaultSendWithTLS_WriteError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(func(conn net.Conn) {
			r := bufio.NewReader(conn)
			_, _ = fmt.Fprintf(conn, "220 localhost ESMTP\r\n")
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					return
				}
				cmd := strings.ToUpper(strings.Fields(strings.TrimRight(line, "\r\n"))[0])
				switch cmd {
				case "EHLO", "HELO":
					_, _ = fmt.Fprintf(conn, "250 localhost\r\n")
				case "MAIL":
					_, _ = fmt.Fprintf(conn, "250 OK\r\n")
				case "RCPT":
					_, _ = fmt.Fprintf(conn, "250 OK\r\n")
				case "DATA":
					_, _ = fmt.Fprintf(conn, "354 Start\r\n")
					_ = conn.Close() // close mid-DATA to trigger write/close error
					return
				}
			}
		}), nil
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected write/close error")
	}
}

func TestDefaultSendWithTLS_AuthError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(func(conn net.Conn) {
			defer func() { _ = conn.Close() }()
			r := bufio.NewReader(conn)
			_, _ = fmt.Fprintf(conn, "220 localhost ESMTP\r\n")
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					return
				}
				cmd := strings.ToUpper(strings.Fields(strings.TrimRight(line, "\r\n"))[0])
				switch cmd {
				case "EHLO", "HELO":
					_, _ = fmt.Fprintf(conn, "250 localhost\r\n")
				case "AUTH":
					_, _ = fmt.Fprintf(conn, "535 5.7.8 Authentication failed\r\n")
					return
				}
			}
		}), nil
	}

	auth := smtp.PlainAuth("", "bad@test.com", "wrongpass", "localhost")
	err := defaultSendWithTLS("localhost:465", auth, "from@test.com", "to@test.com", []byte("msg"))
	if err == nil {
		t.Fatal("expected auth error")
	}
}

func TestDefaultSendWithTLS_LargeMessageWriteError(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(func(conn net.Conn) {
			defer func() { _ = conn.Close() }()
			r := bufio.NewReader(conn)
			_, _ = fmt.Fprintf(conn, "220 localhost ESMTP\r\n")
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					return
				}
				cmd := strings.ToUpper(strings.Fields(strings.TrimRight(line, "\r\n"))[0])
				switch cmd {
				case "EHLO", "HELO":
					_, _ = fmt.Fprintf(conn, "250 localhost\r\n")
				case "MAIL":
					_, _ = fmt.Fprintf(conn, "250 OK\r\n")
				case "RCPT":
					_, _ = fmt.Fprintf(conn, "250 OK\r\n")
				case "DATA":
					_, _ = fmt.Fprintf(conn, "354 Start\r\n")
					_ = conn.Close() // close immediately to fail large write
					return
				}
			}
		}), nil
	}

	// 8KB message forces bufio.Writer buffer flush, triggering write error
	largeMsg := make([]byte, 8192)
	for i := range largeMsg {
		largeMsg[i] = 'x'
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", largeMsg)
	if err == nil {
		t.Fatal("expected write error for large message")
	}
}

func TestDefaultSendWithTLS_Success(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(serveFakeSMTP), nil
	}

	err := defaultSendWithTLS("localhost:465", nil, "from@test.com", "to@test.com", []byte("test message"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDefaultSendWithTLS_WithAuth(t *testing.T) {
	orig := dialTLSFunc
	t.Cleanup(func() { dialTLSFunc = orig })
	dialTLSFunc = func(_, _ string, _ *tls.Config) (net.Conn, error) {
		return withFakeConn(serveFakeSMTP), nil
	}

	auth := smtp.PlainAuth("", "user@test.com", "password", "localhost")
	err := defaultSendWithTLS("localhost:465", auth, "from@test.com", "to@test.com", []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
}
