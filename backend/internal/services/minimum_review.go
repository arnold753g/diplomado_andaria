package services

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"starter-backend/internal/models"
)

// MarkMinimumReviews closes the commercial window without deciding the refund.
// The assigned agency manager must explicitly confirm the cancellation.
func MarkMinimumReviews(db *gorm.DB, now time.Time) (int64, error) {
	result := db.Model(&models.TourPackageDeparture{}).
		Where("estado IN ? AND compras_cierran_en <= ? AND salida_en > ? AND cupos_confirmados < cupo_minimo", []string{"open", "confirmed"}, now, now).
		Updates(map[string]any{
			"estado":             "minimum_review",
			"revision_minimo_en": now.UTC(),
			"es_excepcion":       true,
			"version":            gorm.Expr("version + 1"),
		})
	return result.RowsAffected, result.Error
}

func RunMinimumReviewWorker(ctx context.Context, db *gorm.DB, logger *slog.Logger, interval time.Duration) {
	run := func() {
		count, err := MarkMinimumReviews(db, time.Now())
		if err != nil {
			logger.Error("minimum review evaluation failed", "error", err)
			return
		}
		if count > 0 {
			logger.Info("departures require minimum review", "count", count)
		}
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
