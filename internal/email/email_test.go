package email

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/model/organization"
)

func TestNotificationMIMEPreservesUnicodeAndActionLinks(t *testing.T) {
	sender := NewSender(&config.Config{SMTPFrom: "web@example.org"}, nil)
	n := notification{
		Organization: organization.Details{Name: "Example", ShortName: "Example Council", Institution: "Example University"},
		Name:         "Jörg <script>alert(1)</script>", Domain: "https://example.org",
		Title: "E-Mail-Adresse bestätigen", Intro: "Bestätige dein Konto.",
		Action: "E-Mail-Adresse bestätigen", Link: "https://example.org/email/confirm?token=" + strings.Repeat("a", 64) + "&source=email",
		Validity: "Der Link ist 24 Stunden gültig.", Note: "Deine Fachschaft.",
	}
	encoded, err := sender.message("person@example.org", n)
	if err != nil {
		t.Fatal(err)
	}
	message, err := mail.ReadMessage(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	subject, err := new(mime.WordDecoder).DecodeHeader(message.Header.Get("Subject"))
	if err != nil || subject != n.Title {
		t.Fatal("subject did not preserve German characters")
	}
	kind, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil || kind != "multipart/alternative" {
		t.Fatal("missing alternative email representations")
	}
	parts := multipart.NewReader(message.Body, params["boundary"])
	for _, expected := range []string{"text/plain", "text/html"} {
		part, err := parts.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		kind, params, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil || kind != expected || params["charset"] != "UTF-8" {
			t.Fatal("invalid email part")
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if !strings.Contains(text, n.Title) || !strings.Contains(text, n.Validity) {
			t.Fatal("Unicode message content did not survive transport encoding")
		}
		if expected == "text/plain" {
			if !strings.Contains(text, n.Link) || !strings.Contains(text, n.Name) {
				t.Fatal("plain-text message lost the action link or recipient")
			}
		} else {
			if strings.Contains(text, "<script>") || !strings.Contains(text, "&lt;script&gt;") {
				t.Fatal("recipient name was not escaped")
			}
			if !strings.Contains(text, `href="`+strings.ReplaceAll(n.Link, "&", "&amp;")+`"`) {
				t.Fatal("HTML message lost the action link")
			}
		}
	}
	if _, err := parts.NextPart(); err != io.EOF {
		t.Fatal("unexpected email part")
	}
}
