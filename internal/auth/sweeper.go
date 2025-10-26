package auth

import (
	"context"
	"log"
	"time"

	"github.com/fachschaftinformatik/web/internal/database"
)

func StartSessionSweeper(ctx context.Context, querier database.Querier, logger *log.Logger) {
	logger.Println("Session sweeper started.")
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		if err := querier.DeleteExpiredSessions(ctx); err != nil { //
			logger.Printf("Error sweeping sessions: %v", err)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			logger.Println("Session sweeper stopped.")
			return
		}
	}
}
