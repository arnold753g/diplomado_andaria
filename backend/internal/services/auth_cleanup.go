package services

import (
	"context"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

func RunRegistrationCleanup(ctx context.Context, db *gorm.DB, logger *slog.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		// Pending identity data and password hashes are unnecessary after expiry.
		operation, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := db.WithContext(operation).Exec("DELETE FROM registros_pendientes WHERE vence_en <= ?", time.Now().UTC()).Error
		cancel()
		if err != nil && ctx.Err() == nil {
			logger.Error("expired registration cleanup failed")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
