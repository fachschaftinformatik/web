package email

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"net/url"
	"time"

	"github.com/fachschaftinformatik/web/internal/config"
)

//go:embed templates/*.html
var templateFS embed.FS

type Sender struct {
	cfg *config.Config
	tpl *template.Template
}

func NewSender(cfg *config.Config) *Sender {
	tpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(fmt.Errorf("failed to parse email templates: %w", err))
	}

	return &Sender{
		cfg: cfg,
		tpl: tpl,
	}
}

type verificationData struct {
	Name       string
	VerifyLink string
	Domain     string
}

func (s *Sender) SendVerificationEmail(toEmail, name, link string) error {

	data := verificationData{
		Name:       name,
		VerifyLink: link,
		Domain:     s.cfg.Domain,
	}

	var body bytes.Buffer
	view, subject := "verification.html", "Bitte bestätige deine E-Mail-Adresse"
	if target, err := url.Parse(link); err == nil && target.Path == "/invite" {
		view, subject = "invitation.html", "Einladung zur Fachschaft Informatik"
	}
	if err := s.tpl.ExecuteTemplate(&body, view, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	headers := make(map[string]string)
	headers["From"] = s.cfg.SMTPFrom
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body.String()

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)

	var auth smtp.Auth
	if s.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	}

	if err := send(addr, s.cfg.SMTPHost, auth, s.cfg.SMTPFrom, toEmail, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

type passwordResetData struct {
	Name      string
	ResetLink string
	Domain    string
}

func (s *Sender) SendPasswordResetEmail(toEmail, name, link string) error {

	data := passwordResetData{
		Name:      name,
		ResetLink: link,
		Domain:    s.cfg.Domain,
	}

	var body bytes.Buffer
	if err := s.tpl.ExecuteTemplate(&body, "password_reset.html", data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	subject := "Passwort zurücksetzen"

	headers := make(map[string]string)
	headers["From"] = s.cfg.SMTPFrom
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body.String()

	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)

	var auth smtp.Auth
	if s.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	}

	if err := send(addr, s.cfg.SMTPHost, auth, s.cfg.SMTPFrom, toEmail, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Bound the complete SMTP exchange so unavailable mail cannot block shutdown.
func send(addr, host string, auth smtp.Auth, from, to string, message []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
