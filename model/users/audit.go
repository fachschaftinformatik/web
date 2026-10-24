package users

import (
	"context"
	"log/slog"
	"strings"

	"github.com/fachschaftinformatik/web/internal/audit"
)

func accountEvent(ctx context.Context, event, outcome string, userID int64, email string) {
	_, domain, _ := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@")
	// Rejected addresses can contain line breaks; keep audit details single-line.
	domain = strings.ReplaceAll(domain, "\r", "")
	domain = strings.ReplaceAll(domain, "\n", "")
	audit.Audit(ctx, event, outcome, userID, slog.String("email_domain", domain))
}
