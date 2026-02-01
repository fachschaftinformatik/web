package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"

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

func (s *Sender) SendVerificationEmail(toEmail, name, token string) error {
	link := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", s.cfg.Domain, token)

	data := verificationData{
		Name:       name,
		VerifyLink: link,
		Domain:     s.cfg.Domain,
	}

	var body bytes.Buffer
	if err := s.tpl.ExecuteTemplate(&body, "verification.html", data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	subject := "Bitte bestätige deine E-Mail-Adresse"

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

	if err := smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{toEmail}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

type passwordResetData struct {
	Name      string
	ResetLink string
	Domain    string
}

func (s *Sender) SendPasswordResetEmail(toEmail, name, token string) error {
	link := fmt.Sprintf("%s/auth/reset?token=%s", s.cfg.Domain, token)

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

	if err := smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{toEmail}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
