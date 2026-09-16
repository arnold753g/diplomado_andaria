package handlers

import (
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

type DashboardHandler struct{ db *gorm.DB }

func NewDashboardHandler(db *gorm.DB) *DashboardHandler { return &DashboardHandler{db: db} }

type dashboardNextPurchase struct {
	Reference      string    `json:"reference"`
	PackageName    string    `json:"package_name"`
	DepartureStart time.Time `json:"departure_start"`
	Status         string    `json:"status"`
}

type touristDashboard struct {
	Favorites        int64                  `json:"favorites"`
	ActivePurchases  int64                  `json:"active_purchases"`
	PendingRefunds   int64                  `json:"pending_refunds"`
	CompletedRefunds int64                  `json:"completed_refunds"`
	NextPurchase     *dashboardNextPurchase `json:"next_purchase,omitempty"`
}

type agencyDashboard struct {
	Assigned          bool   `json:"assigned"`
	AgencyName        string `json:"agency_name,omitempty"`
	Packages          int64  `json:"packages"`
	PublishedPackages int64  `json:"published_packages"`
	PaymentsToReview  int64  `json:"payments_to_review"`
	MinimumReviews    int64  `json:"minimum_reviews"`
	PendingRefunds    int64  `json:"pending_refunds"`
	RefundsDueSoon    int64  `json:"refunds_due_soon"`
	OverdueRefunds    int64  `json:"overdue_refunds"`
}

type attractionManagerDashboard struct {
	AssignedAttractions  int64 `json:"assigned_attractions"`
	PublishedAttractions int64 `json:"published_attractions"`
	DraftAttractions     int64 `json:"draft_attractions"`
	InactiveAttractions  int64 `json:"inactive_attractions"`
}

func dashboardError(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, r, http.StatusInternalServerError, "DASHBOARD_ERROR", "No se pudo cargar el resumen de tu cuenta", nil)
}

func (h *DashboardHandler) Mine(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return
	}

	switch principal.User.Role {
	case models.RoleUser:
		h.tourist(w, r, principal.User.ID)
	case models.RoleAgency:
		h.agency(w, r, principal.User.ID)
	case models.RoleAttraction:
		h.attractionManager(w, r, principal.User.ID)
	default:
		respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "Este panel no está disponible para tu rol", nil)
	}
}

func (h *DashboardHandler) tourist(w http.ResponseWriter, r *http.Request, userID uint64) {
	data := touristDashboard{}
	queries := []struct {
		query *gorm.DB
		value *int64
	}{
		{h.db.Model(&models.AttractionFavorite{}).Where("user_id = ?", userID), &data.Favorites},
		{h.db.Model(&models.TourPackagePurchase{}).Where("tourist_id = ? AND status IN ?", userID, []string{models.PurchasePaymentReview, models.PurchaseCorrectionRequested, models.PurchaseConfirmed}), &data.ActivePurchases},
		{h.db.Model(&models.TourPackagePurchase{}).Where("tourist_id = ? AND status = ?", userID, models.PurchaseRefundPending), &data.PendingRefunds},
		{h.db.Model(&models.TourPackagePurchase{}).Where("tourist_id = ? AND status = ?", userID, models.PurchaseRefunded), &data.CompletedRefunds},
	}
	for _, item := range queries {
		if err := item.query.Count(item.value).Error; err != nil {
			dashboardError(w, r)
			return
		}
	}

	var next dashboardNextPurchase
	err := h.db.Table("tour_package_purchases purchase").
		Select("purchase.reference, purchase.status, pkg.name AS package_name, departure.starts_at AS departure_start").
		Joins("JOIN tour_packages pkg ON pkg.id = purchase.package_id").
		Joins("JOIN tour_package_departures departure ON departure.id = purchase.departure_id").
		Where("purchase.tourist_id = ? AND purchase.status IN ? AND departure.starts_at > ?", userID, []string{models.PurchasePaymentReview, models.PurchaseCorrectionRequested, models.PurchaseConfirmed}, time.Now()).
		Order("departure.starts_at, purchase.id").Take(&next).Error
	if err == nil {
		data.NextPurchase = &next
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		dashboardError(w, r)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"role": models.RoleUser, "tourist": data}, "Resumen cargado")
}

func (h *DashboardHandler) agency(w http.ResponseWriter, r *http.Request, userID uint64) {
	data := agencyDashboard{}
	var agency models.Agency
	if err := h.db.Select("id", "name").Where("manager_id = ?", userID).First(&agency).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respond.JSON(w, http.StatusOK, map[string]any{"role": models.RoleAgency, "agency": data}, "Resumen cargado")
			return
		}
		dashboardError(w, r)
		return
	}
	data.Assigned, data.AgencyName = true, agency.Name
	now := time.Now().UTC()
	dueSoon := now.Add(24 * time.Hour)
	queries := []struct {
		query *gorm.DB
		value *int64
	}{
		{h.db.Model(&models.TourPackage{}).Where("agency_id = ?", agency.ID), &data.Packages},
		{h.db.Model(&models.TourPackage{}).Where("agency_id = ? AND published = TRUE", agency.ID), &data.PublishedPackages},
		{h.db.Model(&models.TourPackagePurchase{}).Where("agency_id = ? AND status = ?", agency.ID, models.PurchasePaymentReview), &data.PaymentsToReview},
		{h.db.Model(&models.TourPackagePurchase{}).Where("agency_id = ? AND status = ?", agency.ID, models.PurchaseRefundPending), &data.PendingRefunds},
		{h.db.Model(&models.TourPackagePurchase{}).Where("agency_id = ? AND status = ? AND refund_due_at > ? AND refund_due_at <= ?", agency.ID, models.PurchaseRefundPending, now, dueSoon), &data.RefundsDueSoon},
		{h.db.Model(&models.TourPackagePurchase{}).Where("agency_id = ? AND status = ? AND refund_due_at <= ?", agency.ID, models.PurchaseRefundPending, now), &data.OverdueRefunds},
		{h.db.Model(&models.TourPackageDeparture{}).Joins("JOIN tour_packages pkg ON pkg.id = tour_package_departures.package_id").Where("pkg.agency_id = ? AND tour_package_departures.status = ?", agency.ID, "minimum_review"), &data.MinimumReviews},
	}
	for _, item := range queries {
		if err := item.query.Count(item.value).Error; err != nil {
			dashboardError(w, r)
			return
		}
	}
	respond.JSON(w, http.StatusOK, map[string]any{"role": models.RoleAgency, "agency": data}, "Resumen cargado")
}

func (h *DashboardHandler) attractionManager(w http.ResponseWriter, r *http.Request, userID uint64) {
	data := attractionManagerDashboard{}
	queries := []struct {
		query *gorm.DB
		value *int64
	}{
		{h.db.Model(&models.Attraction{}).Where("manager_id = ?", userID), &data.AssignedAttractions},
		{h.db.Model(&models.Attraction{}).Where("manager_id = ? AND status = ? AND published = TRUE", userID, models.StatusActive), &data.PublishedAttractions},
		{h.db.Model(&models.Attraction{}).Where("manager_id = ? AND status = ? AND published = FALSE", userID, models.StatusActive), &data.DraftAttractions},
		{h.db.Model(&models.Attraction{}).Where("manager_id = ? AND status = ?", userID, models.StatusInactive), &data.InactiveAttractions},
	}
	for _, item := range queries {
		if err := item.query.Count(item.value).Error; err != nil {
			dashboardError(w, r)
			return
		}
	}
	respond.JSON(w, http.StatusOK, map[string]any{"role": models.RoleAttraction, "attraction_manager": data}, "Resumen cargado")
}
