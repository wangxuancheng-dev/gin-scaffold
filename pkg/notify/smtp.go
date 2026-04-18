package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"gin-scaffold/config"
)

// SMTPNotifier 通过 SMTP 发送纯文本邮件。
type SMTPNotifier struct {
	cfg config.SMTPNotifyConfig
}

// NewSMTPNotifier 构造 SMTP 通知器。
func NewSMTPNotifier(cfg config.SMTPNotifyConfig) *SMTPNotifier {
	return &SMTPNotifier{cfg: cfg}
}

// Notify 发送邮件；Meta["to"] 可覆盖默认收件人。
func (n *SMTPNotifier) Notify(ctx context.Context, msg Message) error {
	if n == nil {
		return fmt.Errorf("notify: smtp: nil notifier")
	}
	to := strings.TrimSpace(msg.Meta["to"])
	if to == "" {
		to = strings.TrimSpace(n.cfg.ToDefault)
	}
	if to == "" {
		return fmt.Errorf("notify: smtp: missing recipient")
	}
	subject := strings.TrimSpace(msg.Title)
	if subject == "" {
		subject = "notification"
	}
	body := msg.Body
	addr := fmt.Sprintf("%s:%d", strings.TrimSpace(n.cfg.Host), n.cfg.Port)
	from := strings.TrimSpace(n.cfg.From)

	var auth smtp.Auth
	if strings.TrimSpace(n.cfg.Username) != "" || strings.TrimSpace(n.cfg.Password) != "" {
		auth = smtp.PlainAuth("", strings.TrimSpace(n.cfg.Username), n.cfg.Password, strings.TrimSpace(n.cfg.Host))
	}

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n", from, to, subject)
	raw := []byte(headers + body)

	dialer := net.Dialer{Timeout: 15 * time.Second}
	if dl, ok := ctx.Deadline(); ok {
		if d := time.Until(dl); d > 0 {
			dialer.Timeout = d
		}
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("notify: smtp dial: %w", err)
	}
	defer conn.Close()

	var client *smtp.Client
	if n.cfg.ImplicitTLS {
		tlsCfg := &tls.Config{
			ServerName:         strings.TrimSpace(n.cfg.Host),
			InsecureSkipVerify: n.cfg.SkipVerify,
			MinVersion:         tls.VersionTLS12,
		}
		tlsConn := tls.Client(conn, tlsCfg)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("notify: smtp tls handshake: %w", err)
		}
		client, err = smtp.NewClient(tlsConn, strings.TrimSpace(n.cfg.Host))
	} else {
		client, err = smtp.NewClient(conn, strings.TrimSpace(n.cfg.Host))
	}
	if err != nil {
		return fmt.Errorf("notify: smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if !n.cfg.ImplicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsCfg := &tls.Config{
				ServerName:         strings.TrimSpace(n.cfg.Host),
				InsecureSkipVerify: n.cfg.SkipVerify,
				MinVersion:         tls.VersionTLS12,
			}
			if err := client.StartTLS(tlsCfg); err != nil {
				return fmt.Errorf("notify: smtp starttls: %w", err)
			}
		}
	}
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("notify: smtp auth: %w", err)
			}
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("notify: smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("notify: smtp rcpt: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("notify: smtp data: %w", err)
	}
	if _, err := wc.Write(raw); err != nil {
		_ = wc.Close()
		return fmt.Errorf("notify: smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("notify: smtp close writer: %w", err)
	}
	return client.Quit()
}
