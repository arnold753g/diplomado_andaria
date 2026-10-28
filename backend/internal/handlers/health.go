package handlers

import (
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"

	"starter-backend/internal/respond"
)

// NewHealth checks database readiness, including the database connection used by the API.
func NewHealth(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		var value int
		if db == nil || db.WithContext(ctx).Raw("SELECT 1").Scan(&value).Error != nil || value != 1 {
			respond.Error(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Servicio temporalmente no disponible.", nil)
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"}, "")
	}
}
