package handler

import (
	"context"
	"time"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/go-chi/httplog/v2"
)

func StartSessionSweeper(ctx context.Context, querier database.Querier, logger *httplog.Logger) {
	logger.Info("Session sweeper started.")
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		if err := querier.DeleteExpiredSessions(ctx); err != nil {
			logger.Error("Error sweeping sessions", "err", err)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			logger.Info("Session sweeper stopped.")
			return
		}
	}
}
