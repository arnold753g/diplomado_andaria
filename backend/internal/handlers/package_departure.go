package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

var (
	errDepartureConflict = errors.New("departure version conflict")
	errDepartureFinal    = errors.New("departure is in a final state")
)

type departureUpdateInput struct {
	Version            int    `json:"version"`
	MeetingTime        string `json:"meeting_time"`
	MinCapacity        int    `json:"min_capacity"`
	MaxCapacity        int    `json:"max_capacity"`
	BookingCutoffHours int    `json:"booking_cutoff_hours"`
	MeetingPoint       string `json:"meeting_point"`
	Instructions       string `json:"instructions"`
}

type departureCancelInput struct {
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}

type minimumRefundInput struct {
	Version int `json:"version"`
}

func cancelDeparturePurchases(tx *gorm.DB, departure *models.TourPackageDeparture, userID uint64, reason string, now time.Time) error {
	due := now.Add(refundProcessingWindow)
	if err := tx.Model(&models.TourPackagePurchase{}).
		Where("departure_id = ? AND status IN ?", departure.ID, []string{models.PurchasePaymentReview, models.PurchaseCorrectionRequested, models.PurchaseConfirmed}).
		Updates(map[string]any{
			"status": models.PurchaseRefundPending, "rejection_reason": "", "refund_reason": reason,
			"refund_requested_at": now, "refund_due_at": due, "cancellation_reason": reason, "cancelled_at": now,
			"agency_review_user_id": userID, "reviewed_at": now, "version": gorm.Expr("version + 1"),
		}).Error; err != nil {
		return err
	}
	departure.HeldCapacity = 0
	departure.ConfirmedCapacity = 0
	return nil
}

func departurePathID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(mux.Vars(r)["departure"], 10, 64)
	if err != nil || id == 0 {
		respond.Error(w, r, http.StatusNotFound, "DEPARTURE_NOT_FOUND", "La salida no está disponible", nil)
		return 0, false
	}
	return id, true
}

func validateDepartureUpdate(input *departureUpdateInput, departure models.TourPackageDeparture, now time.Time) (time.Time, time.Time, map[string]string) {
	fields := map[string]string{}
	if input.Version < 1 {
		fields["version"] = "Recarga la salida antes de modificarla"
	}
	meetingClock, meetingHour, meetingMinute, meetingErr := normalizeClock(input.MeetingTime)
	if meetingErr != nil {
		fields["meeting_time"] = "Selecciona una hora de encuentro válida"
	} else {
		input.MeetingTime = meetingClock
	}
	localStart := departure.StartsAt.In(boliviaTime)
	meetingAt := time.Date(localStart.Year(), localStart.Month(), localStart.Day(), meetingHour, meetingMinute, 0, 0, boliviaTime)
	if meetingErr == nil && meetingAt.After(departure.StartsAt) {
		fields["meeting_time"] = "La hora de encuentro no puede ser posterior a la salida"
	}
	if input.MinCapacity < 1 || input.MinCapacity > 500 || input.MaxCapacity < input.MinCapacity || input.MaxCapacity > 500 {
		fields["capacity"] = "Define cupos válidos entre 1 y 500"
	} else if input.MaxCapacity < departure.HeldCapacity+departure.ConfirmedCapacity {
		fields["max_capacity"] = "El cupo máximo no puede ser menor que los cupos ocupados"
	}
	if input.BookingCutoffHours < 1 || input.BookingCutoffHours > 720 {
		fields["booking_cutoff_hours"] = "El cierre debe estar entre 1 y 720 horas antes de la salida"
	}
	bookingClosesAt := departure.StartsAt.Add(-time.Duration(input.BookingCutoffHours) * time.Hour)
	if !bookingClosesAt.After(departure.BookingOpensAt) {
		fields["booking_cutoff_hours"] = "El cierre debe ser posterior a la apertura de compras"
	}
	if !bookingClosesAt.After(now) {
		fields["booking_cutoff_hours"] = "El nuevo plazo de compra ya habría terminado"
	}
	input.MeetingPoint = strings.TrimSpace(input.MeetingPoint)
	input.Instructions = strings.TrimSpace(input.Instructions)
	if len([]rune(input.MeetingPoint)) > 500 {
		fields["meeting_point"] = "El punto de encuentro admite hasta 500 caracteres"
	}
	if len([]rune(input.Instructions)) > 3000 {
		fields["instructions"] = "Las instrucciones admiten hasta 3000 caracteres"
	}
	return meetingAt, bookingClosesAt, fields
}

func (h *PackageHandler) UpdateDeparture(w http.ResponseWriter, r *http.Request) {
	packageID, ok := pathID(w, r)
	if !ok {
		return
	}
	departureID, ok := departurePathID(w, r)
	if !ok {
		return
	}
	var input departureUpdateInput
	if !decodeJSONLimit(w, r, &input, 32<<10) {
		return
	}
	var departure models.TourPackageDeparture
	err := h.db.Transaction(func(tx *gorm.DB) error {
		agency, err := h.ownAgency(r, tx)
		if err != nil {
			return err
		}
		if agency.Status != models.StatusActive {
			return errPackageAgencyInactive
		}
		var item models.TourPackage
		if err := tx.Select("id").Where("id = ? AND agency_id = ?", packageID, agency.ID).First(&item).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND package_id = ?", departureID, packageID).First(&departure).Error; err != nil {
			return err
		}
		if departure.Version != input.Version {
			return errDepartureConflict
		}
		if departure.Status == "cancelled" || departure.Status == "completed" || !departure.StartsAt.After(time.Now()) {
			return errDepartureFinal
		}
		meetingAt, bookingClosesAt, fields := validateDepartureUpdate(&input, departure, time.Now())
		if len(fields) > 0 {
			return &departureValidationError{fields: fields}
		}
		principal, _ := middleware.CurrentPrincipal(r)
		departure.MeetingAt = meetingAt
		departure.BookingClosesAt = bookingClosesAt
		departure.MinCapacity = input.MinCapacity
		departure.MaxCapacity = input.MaxCapacity
		departure.MeetingPoint = input.MeetingPoint
		departure.Instructions = input.Instructions
		departure.IsException = true
		departure.ModifiedByUserID = &principal.User.ID
		departure.Version++
		if departure.ConfirmedCapacity >= departure.MinCapacity {
			departure.Status = "confirmed"
		} else {
			departure.Status = "open"
		}
		return tx.Save(&departure).Error
	})
	if err != nil {
		departureError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, departure, "Salida actualizada")
}

func (h *PackageHandler) CancelDeparture(w http.ResponseWriter, r *http.Request) {
	packageID, ok := pathID(w, r)
	if !ok {
		return
	}
	departureID, ok := departurePathID(w, r)
	if !ok {
		return
	}
	var input departureCancelInput
	if !decodeJSONLimit(w, r, &input, 16<<10) {
		return
	}
	input.Reason = strings.TrimSpace(input.Reason)
	fields := map[string]string{}
	if input.Version < 1 {
		fields["version"] = "Recarga la salida antes de cancelarla"
	}
	if len([]rune(input.Reason)) < 5 || len([]rune(input.Reason)) > 1000 {
		fields["reason"] = "Explica el motivo de cancelación con entre 5 y 1000 caracteres"
	}
	if len(fields) > 0 {
		respond.Error(w, r, http.StatusUnprocessableEntity, "DEPARTURE_VALIDATION", "Revisa los datos de la salida", fields)
		return
	}
	var departure models.TourPackageDeparture
	err := h.db.Transaction(func(tx *gorm.DB) error {
		agency, err := h.ownAgency(r, tx)
		if err != nil {
			return err
		}
		if agency.Status != models.StatusActive {
			return errPackageAgencyInactive
		}
		var item models.TourPackage
		if err := tx.Select("id").Where("id = ? AND agency_id = ?", packageID, agency.ID).First(&item).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND package_id = ?", departureID, packageID).First(&departure).Error; err != nil {
			return err
		}
		if departure.Version != input.Version {
			return errDepartureConflict
		}
		if departure.Status == "cancelled" || departure.Status == "completed" || !departure.StartsAt.After(time.Now()) {
			return errDepartureFinal
		}
		principal, _ := middleware.CurrentPrincipal(r)
		now := time.Now()
		departure.Status = "cancelled"
		departure.CancellationReason = input.Reason
		departure.CancelledAt = &now
		departure.IsException = true
		departure.ModifiedByUserID = &principal.User.ID
		if err := cancelDeparturePurchases(tx, &departure, principal.User.ID, "Salida cancelada: "+input.Reason, now); err != nil {
			return err
		}
		departure.Version++
		return tx.Save(&departure).Error
	})
	if err != nil {
		departureError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, departure, "Salida cancelada")
}

func (h *PackageHandler) ConfirmMinimumRefund(w http.ResponseWriter, r *http.Request) {
	packageID, ok := pathID(w, r)
	if !ok {
		return
	}
	departureID, ok := departurePathID(w, r)
	if !ok {
		return
	}
	var input minimumRefundInput
	if !decodeJSONLimit(w, r, &input, 8<<10) {
		return
	}
	if input.Version < 1 {
		respond.Error(w, r, http.StatusUnprocessableEntity, "DEPARTURE_VALIDATION", "Recarga la salida antes de confirmar", nil)
		return
	}
	var departure models.TourPackageDeparture
	err := h.db.Transaction(func(tx *gorm.DB) error {
		agency, err := h.ownAgency(r, tx)
		if err != nil {
			return err
		}
		if agency.Status != models.StatusActive {
			return errPackageAgencyInactive
		}
		var item models.TourPackage
		if err := tx.Select("id").Where("id = ? AND agency_id = ?", packageID, agency.ID).First(&item).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND package_id = ?", departureID, packageID).First(&departure).Error; err != nil {
			return err
		}
		if departure.Version != input.Version {
			return errDepartureConflict
		}
		now := time.Now().UTC()
		if departure.Status != "minimum_review" || departure.BookingClosesAt.After(now) || !departure.StartsAt.After(now) || departure.ConfirmedCapacity >= departure.MinCapacity {
			return errDepartureFinal
		}
		principal, _ := middleware.CurrentPrincipal(r)
		reason := "No se alcanzó el cupo mínimo antes del cierre de compras"
		departure.Status = "cancelled"
		departure.CancellationReason = reason
		departure.CancelledAt = &now
		departure.MinimumReviewedAt = &now
		departure.MinimumReviewUserID = &principal.User.ID
		departure.IsException = true
		departure.ModifiedByUserID = &principal.User.ID
		if err := cancelDeparturePurchases(tx, &departure, principal.User.ID, reason, now); err != nil {
			return err
		}
		departure.Version++
		return tx.Save(&departure).Error
	})
	if err != nil {
		departureError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, departure, "Salida cancelada y reembolsos iniciados")
}

type departureValidationError struct{ fields map[string]string }

func (e *departureValidationError) Error() string { return "departure validation failed" }

func departureError(w http.ResponseWriter, r *http.Request, err error) {
	var validation *departureValidationError
	switch {
	case errors.As(err, &validation):
		respond.Error(w, r, http.StatusUnprocessableEntity, "DEPARTURE_VALIDATION", "Revisa los datos de la salida", validation.fields)
	case errors.Is(err, gorm.ErrRecordNotFound):
		respond.Error(w, r, http.StatusNotFound, "DEPARTURE_NOT_FOUND", "La salida no está disponible o no tienes acceso", nil)
	case errors.Is(err, errDepartureConflict):
		respond.Error(w, r, http.StatusConflict, "DEPARTURE_CONFLICT", "La salida cambió desde que la abriste. Recarga antes de continuar", nil)
	case errors.Is(err, errDepartureFinal):
		respond.Error(w, r, http.StatusUnprocessableEntity, "DEPARTURE_FINAL", "Una salida cancelada, realizada o pasada ya no puede modificarse", nil)
	case errors.Is(err, errPackageAgencyInactive):
		respond.Error(w, r, http.StatusForbidden, "PACKAGE_AGENCY_INACTIVE", "La agencia está desactivada. Contacta al administrador", nil)
	default:
		respond.Error(w, r, http.StatusInternalServerError, "DEPARTURE_ERROR", "No se pudo completar la operación de la salida", nil)
	}
}
