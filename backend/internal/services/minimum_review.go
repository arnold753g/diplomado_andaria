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
		Where("status IN ? AND booking_closes_at <= ? AND starts_at > ? AND confirmed_capacity < min_capacity", []string{"open", "confirmed"}, now, now).
		Updates(map[string]any{
			"status":            "minimum_review",
			"minimum_review_at": now.UTC(),
			"is_exception":      true,
			"version":           gorm.Expr("version + 1"),
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
