package notifications

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/admin"
)

// ErrNotConfigured is returned when SMTP is not configured.
var ErrNotConfigured = errors.New("smtp not configured")

// Mailer sends plain-text email via configured SMTP relay.
type Mailer struct {
	Config admin.SMTPConfig
}

// Enabled reports whether the mailer can send messages.
func (m *Mailer) Enabled() bool {
	return m.Config.Enabled()
}

// Send delivers a plain-text email to one recipient.
func (m *Mailer) Send(_ context.Context, to, subject, body string) error {
	if !m.Enabled() {
		return ErrNotConfigured
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return errors.New("destinataire requis")
	}

	from := strings.TrimSpace(m.Config.From)
	msg := buildMessage(from, to, subject, body)
	addr := net.JoinHostPort(m.Config.Host, strconv.Itoa(m.Config.Port))

	var auth smtp.Auth
	if strings.TrimSpace(m.Config.Username) != "" {
		auth = smtp.PlainAuth("", m.Config.Username, m.Config.Password, m.Config.Host)
	}

	if !m.Config.TLS {
		return smtp.SendMail(addr, auth, from, []string{to}, msg)
	}
	if m.Config.Port == 465 {
		return sendImplicitTLS(addr, m.Config.Host, auth, from, []string{to}, msg)
	}

	return sendSTARTTLS(addr, m.Config.Host, auth, from, []string{to}, msg)
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(from)
	b.WriteString("\r\nTo: ")
	b.WriteString(to)
	b.WriteString("\r\nSubject: ")
	b.WriteString(subject)
	b.WriteString("\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func sendImplicitTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	return sendWithClient(client, auth, from, to, msg)
}

func sendSTARTTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	return sendWithClient(client, auth, from, to, msg)
}

func sendWithClient(client *smtp.Client, auth smtp.Auth, from string, to []string, msg []byte) error {
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}
