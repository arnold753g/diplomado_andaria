package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

var (
	errPurchaseUnavailable  = errors.New("purchase departure unavailable")
	errPurchaseCapacity     = errors.New("purchase capacity unavailable")
	errPurchasePayment      = errors.New("purchase payment method unavailable")
	errPurchaseConflict     = errors.New("purchase version conflict")
	errPurchaseState        = errors.New("purchase state conflict")
	errPurchaseCancellation = errors.New("purchase cancellation unavailable")
	errRefundDestination    = errors.New("refund destination invalid")
)

const refundProcessingWindow = 72 * time.Hour

type PurchaseHandler struct{ db *gorm.DB }

func NewPurchaseHandler(db *gorm.DB) *PurchaseHandler { return &PurchaseHandler{db: db} }

type purchaseMinorInput struct {
	Age       int  `json:"age"`
	IsForeign bool `json:"is_foreign"`
}

type purchaseCreateInput struct {
	DepartureID    uint64               `json:"departure_id"`
	PaymentMethod  string               `json:"payment_method"`
	NationalAdults int                  `json:"national_adults"`
	ForeignAdults  int                  `json:"foreign_adults"`
	Minors         []purchaseMinorInput `json:"minors"`
	PaymentProof   string               `json:"payment_proof"`
}

type purchaseReviewInput struct {
	Version  int    `json:"version"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type purchaseProofInput struct {
	Version      int    `json:"version"`
	PaymentProof string `json:"payment_proof"`
}

type refundDestinationInput struct {
	Version             int    `json:"version"`
	RefundMethod        string `json:"refund_method"`
	RefundQR            string `json:"refund_qr"`
	RefundBankName      string `json:"refund_bank_name"`
	RefundAccountHolder string `json:"refund_account_holder"`
	RefundAccountNumber string `json:"refund_account_number"`
}

type purchaseCancelInput struct {
	refundDestinationInput
	Reason string `json:"reason"`
}

type refundCompleteInput struct {
	Version         int    `json:"version"`
	RefundProof     string `json:"refund_proof"`
	RefundReference string `json:"refund_reference"`
}

type purchasePaymentOptions struct {
	PackageID           uint64   `json:"package_id"`
	AgencyName          string   `json:"agency_name"`
	MinimumPayingAge    int      `json:"minimum_paying_age"`
	Methods             []string `json:"methods"`
	BankName            string   `json:"bank_name"`
	AccountHolder       string   `json:"account_holder"`
	AccountNumber       string   `json:"account_number"`
	PaymentInstructions string   `json:"payment_instructions"`
	QRImage             string   `json:"qr_image,omitempty"`
}

type purchaseResponse struct {
	models.TourPackagePurchase
	PackageName          string     `json:"package_name"`
	AgencyName           string     `json:"agency_name"`
	DepartureStart       time.Time  `json:"departure_start"`
	MeetingPoint         string     `json:"meeting_point"`
	HasProof             bool       `json:"has_proof"`
	HasRefundQR          bool       `json:"has_refund_qr"`
	HasRefundProof       bool       `json:"has_refund_proof"`
	Cancellable          bool       `json:"cancellable"`
	CancellationDeadline *time.Time `json:"cancellation_deadline,omitempty"`
	RefundOverdue        bool       `json:"refund_overdue"`
}

func purchaseQuery(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Package", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "name", "cancellation_allowed", "cancellation_notice_hours")
		}).
		Preload("Agency", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "name") }).
		Preload("Departure", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "starts_at", "meeting_point") })
}

func purchaseView(item models.TourPackagePurchase) purchaseResponse {
	now := time.Now()
	view := purchaseResponse{
		TourPackagePurchase: item,
		HasProof:            len(item.PaymentProof) > 0,
		HasRefundQR:         item.RefundMethod == "qr",
		HasRefundProof:      item.RefundedAt != nil,
		RefundOverdue:       item.Status == models.PurchaseRefundPending && item.RefundDueAt != nil && !item.RefundDueAt.After(now),
	}
	view.PaymentProof = nil
	view.RefundQR = nil
	view.RefundProof = nil
	if item.Package != nil {
		view.PackageName = item.Package.Name
	}
	if item.Agency != nil {
		view.AgencyName = item.Agency.Name
	}
	if item.Departure != nil {
		view.DepartureStart = item.Departure.StartsAt
		view.MeetingPoint = item.Departure.MeetingPoint
	}
	if item.Package != nil && item.Departure != nil && item.Package.CancellationAllowed {
		deadline := item.Departure.StartsAt.Add(-time.Duration(item.Package.CancellationNoticeHours) * time.Hour)
		view.CancellationDeadline = &deadline
		view.Cancellable = deadline.After(now) && item.Departure.StartsAt.After(now) && slices.Contains([]string{models.PurchasePaymentReview, models.PurchaseCorrectionRequested, models.PurchaseConfirmed}, item.Status)
	}
	return view
}

func purchaseReference() (string, error) {
	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("AND-%s-%s", time.Now().UTC().Format("20060102"), strings.ToUpper(hex.EncodeToString(suffix[:]))), nil
}

func (h *PurchaseHandler) PaymentOptions(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var item models.TourPackage
	if err := packagePublicScope(h.db).Preload("Agency").First(&item, id).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	methods := make([]string, 0, 2)
	if item.Agency.AcceptsQR && len(item.Agency.QRImage) > 0 {
		methods = append(methods, "qr")
	}
	if item.Agency.AcceptsTransfer {
		methods = append(methods, "transfer")
	}
	options := purchasePaymentOptions{
		PackageID: item.ID, AgencyName: item.Agency.Name, MinimumPayingAge: item.Agency.MinimumPayingAge,
		Methods: methods, BankName: item.Agency.BankName, AccountHolder: item.Agency.AccountHolder,
		AccountNumber: item.Agency.AccountNumber, PaymentInstructions: item.Agency.PaymentInstructions,
	}
	if slices.Contains(methods, "qr") {
		options.QRImage = "data:image/png;base64," + base64.StdEncoding.EncodeToString(item.Agency.QRImage)
	}
	respond.JSON(w, http.StatusOK, options, "")
}

func (h *PurchaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	var input purchaseCreateInput
	if !decodeJSONLimit(w, r, &input, 7<<20) {
		return
	}
	input.PaymentMethod = strings.TrimSpace(input.PaymentMethod)
	fields := map[string]string{}
	if input.DepartureID == 0 {
		fields["departure_id"] = "Selecciona una salida"
	}
	if !slices.Contains([]string{"qr", "transfer"}, input.PaymentMethod) {
		fields["payment_method"] = "Selecciona QR o transferencia"
	}
	if input.NationalAdults < 0 || input.ForeignAdults < 0 || input.NationalAdults+input.ForeignAdults < 1 || input.NationalAdults+input.ForeignAdults > 100 {
		fields["adults"] = "La compra requiere entre 1 y 100 adultos"
	}
	if len(input.Minors) > 100 {
		fields["minors"] = "Puedes registrar hasta 100 menores"
	}
	for _, minor := range input.Minors {
		if minor.Age < 0 || minor.Age > 17 {
			fields["minors"] = "Las edades de menores deben estar entre 0 y 17 años"
			break
		}
	}
	proof, err := decodeAttractionImage(input.PaymentProof)
	if err != nil {
		fields["payment_proof"] = "Sube un comprobante PNG o JPG de hasta 5 MB"
	}
	if len(fields) > 0 {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PURCHASE_VALIDATION", "Revisa los datos de la compra", fields)
		return
	}

	var purchase models.TourPackagePurchase
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var departure models.TourPackageDeparture
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&departure, input.DepartureID).Error; err != nil {
			return err
		}
		var item models.TourPackage
		if err := packagePublicScope(tx).Preload("Agency").First(&item, departure.PackageID).Error; err != nil {
			return errPurchaseUnavailable
		}
		now := time.Now()
		if !slices.Contains([]string{"open", "confirmed"}, departure.Status) || !departure.StartsAt.After(now) || now.Before(departure.BookingOpensAt) || !now.Before(departure.BookingClosesAt) {
			return errPurchaseUnavailable
		}
		if (input.PaymentMethod == "qr" && (!item.Agency.AcceptsQR || len(item.Agency.QRImage) == 0)) || (input.PaymentMethod == "transfer" && !item.Agency.AcceptsTransfer) {
			return errPurchasePayment
		}
		minors := make([]models.PurchaseMinor, 0, len(input.Minors))
		freeMinors, payingMinors, foreignPayingMinors := 0, 0, 0
		for _, minor := range input.Minors {
			pays := minor.Age >= item.Agency.MinimumPayingAge
			minors = append(minors, models.PurchaseMinor{Age: minor.Age, IsForeign: minor.IsForeign, Pays: pays})
			if pays {
				payingMinors++
				if minor.IsForeign {
					foreignPayingMinors++
				}
			} else {
				freeMinors++
			}
		}
		capacity := input.NationalAdults + input.ForeignAdults + payingMinors
		if capacity < 1 || departure.HeldCapacity+departure.ConfirmedCapacity+capacity > departure.MaxCapacity {
			return errPurchaseCapacity
		}
		total := capacity*item.NationalPriceCents + (input.ForeignAdults+foreignPayingMinors)*item.ForeignSurchargeCents
		if total <= 0 {
			return errPurchaseUnavailable
		}
		reference, err := purchaseReference()
		if err != nil {
			return err
		}
		purchase = models.TourPackagePurchase{
			Reference: reference, TouristID: principal.User.ID, AgencyID: item.AgencyID, PackageID: item.ID, DepartureID: departure.ID,
			Status: models.PurchasePaymentReview, PaymentMethod: input.PaymentMethod,
			NationalAdults: input.NationalAdults, ForeignAdults: input.ForeignAdults, Minors: minors,
			FreeMinorCount: freeMinors, PayingMinorCount: payingMinors, CapacityCount: capacity,
			NationalUnitPriceCents: item.NationalPriceCents, ForeignSurchargeCents: item.ForeignSurchargeCents, TotalCents: total,
			TouristName: strings.TrimSpace(principal.User.FirstName + " " + principal.User.LastName), TouristEmail: principal.User.Email,
			TouristPhone: principal.User.Phone, TouristDocument: principal.User.DocumentNumber, TouristNationality: principal.User.Nationality,
			PaymentProof: proof, Version: 1,
		}
		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}
		departure.HeldCapacity += capacity
		departure.Version++
		return tx.Save(&departure).Error
	})
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	if err := purchaseQuery(h.db).First(&purchase, purchase.ID).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusCreated, purchaseView(purchase), "Compra registrada y pago enviado a revisión")
}

func (h *PurchaseHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	h.list(w, r, h.db.Where("tourist_id = ?", principal.User.ID))
}

func (h *PurchaseHandler) ListAgency(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	q := h.db.Where("agency_id = ?", agency.ID)
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		if !slices.Contains([]string{models.PurchasePaymentReview, models.PurchaseCorrectionRequested, models.PurchaseConfirmed, models.PurchasePaymentRejected, models.PurchaseRefundPending, models.PurchaseRefunded}, status) {
			respond.Error(w, r, http.StatusUnprocessableEntity, "PURCHASE_FILTER_INVALID", "Selecciona un estado válido", nil)
			return
		}
		q = q.Where("status = ?", status)
	}
	h.list(w, r, q)
}

func (h *PurchaseHandler) list(w http.ResponseWriter, r *http.Request, q *gorm.DB) {
	page, limit := queryInt(r, "page", 1, 1, 100000), queryInt(r, "limit", 20, 1, 50)
	var total int64
	if err := q.Model(&models.TourPackagePurchase{}).Count(&total).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	var purchases []models.TourPackagePurchase
	if err := purchaseQuery(q).Omit("payment_proof", "refund_qr", "refund_proof").Order("tour_package_purchases.created_at DESC, tour_package_purchases.id DESC").Limit(limit).Offset((page - 1) * limit).Find(&purchases).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	views := make([]purchaseResponse, 0, len(purchases))
	for _, item := range purchases {
		view := purchaseView(item)
		view.HasProof = true
		views = append(views, view)
	}
	respond.JSON(w, http.StatusOK, map[string]any{"purchases": views, "pagination": map[string]any{"page": page, "limit": limit, "total": total}}, "")
}

func (h *PurchaseHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	h.get(w, r, h.db.Where("tourist_id = ?", principal.User.ID))
}

func (h *PurchaseHandler) GetAgency(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	h.get(w, r, h.db.Where("agency_id = ?", agency.ID))
}

func (h *PurchaseHandler) get(w http.ResponseWriter, r *http.Request, q *gorm.DB) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var purchase models.TourPackagePurchase
	if err := purchaseQuery(q).Omit("payment_proof", "refund_qr", "refund_proof").First(&purchase, id).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	view := purchaseView(purchase)
	view.HasProof = true
	respond.JSON(w, http.StatusOK, view, "")
}

func (h *PurchaseHandler) ProofMine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	h.proof(w, r, h.db.Where("tourist_id = ?", principal.User.ID))
}

func (h *PurchaseHandler) ProofAgency(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	h.proof(w, r, h.db.Where("agency_id = ?", agency.ID))
}

func (h *PurchaseHandler) UpdateProof(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input purchaseProofInput
	if !decodeJSONLimit(w, r, &input, 7<<20) {
		return
	}
	proof, err := decodeAttractionImage(input.PaymentProof)
	if input.Version < 1 || err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PURCHASE_PROOF_INVALID", "Sube un comprobante PNG o JPG de hasta 5 MB", nil)
		return
	}
	var purchase models.TourPackagePurchase
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("tourist_id = ?", principal.User.ID).First(&purchase, id).Error; err != nil {
			return err
		}
		if purchase.Version != input.Version {
			return errPurchaseConflict
		}
		if purchase.Status != models.PurchaseCorrectionRequested {
			return errPurchaseState
		}
		purchase.PaymentProof = proof
		purchase.Status = models.PurchasePaymentReview
		purchase.AgencyReviewUserID = nil
		purchase.ReviewedAt = nil
		purchase.RejectionReason = ""
		purchase.Version++
		return tx.Save(&purchase).Error
	})
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	if err := purchaseQuery(h.db).First(&purchase, purchase.ID).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, purchaseView(purchase), "Nuevo comprobante enviado a revisión")
}

func validateRefundDestination(input *refundDestinationInput) ([]byte, error) {
	input.RefundMethod = strings.TrimSpace(input.RefundMethod)
	input.RefundBankName = strings.TrimSpace(input.RefundBankName)
	input.RefundAccountHolder = strings.TrimSpace(input.RefundAccountHolder)
	input.RefundAccountNumber = strings.TrimSpace(input.RefundAccountNumber)
	if input.RefundMethod == "qr" {
		qr, err := decodeAttractionImage(input.RefundQR)
		if err != nil {
			return nil, errRefundDestination
		}
		return qr, nil
	}
	if input.RefundMethod != "bank_transfer" || len([]rune(input.RefundBankName)) < 2 || len([]rune(input.RefundBankName)) > 120 || len([]rune(input.RefundAccountHolder)) < 3 || len([]rune(input.RefundAccountHolder)) > 160 || len([]rune(input.RefundAccountNumber)) < 3 || len([]rune(input.RefundAccountNumber)) > 80 {
		return nil, errRefundDestination
	}
	return nil, nil
}

func applyRefundDestination(purchase *models.TourPackagePurchase, input refundDestinationInput, qr []byte) {
	purchase.RefundMethod = input.RefundMethod
	purchase.RefundQR = qr
	purchase.RefundBankName = ""
	purchase.RefundAccountHolder = ""
	purchase.RefundAccountNumber = ""
	if input.RefundMethod == "bank_transfer" {
		purchase.RefundBankName = input.RefundBankName
		purchase.RefundAccountHolder = input.RefundAccountHolder
		purchase.RefundAccountNumber = input.RefundAccountNumber
	}
}

func beginPurchaseRefund(purchase *models.TourPackagePurchase, reason string, now time.Time) {
	due := now.Add(refundProcessingWindow)
	purchase.Status = models.PurchaseRefundPending
	purchase.RefundReason = reason
	purchase.RefundRequestedAt = &now
	purchase.RefundDueAt = &due
	purchase.RejectionReason = ""
}

func releasePurchaseCapacity(departure *models.TourPackageDeparture, purchase models.TourPackagePurchase) error {
	switch purchase.Status {
	case models.PurchasePaymentReview, models.PurchaseCorrectionRequested:
		if departure.HeldCapacity < purchase.CapacityCount {
			return errPurchaseState
		}
		departure.HeldCapacity -= purchase.CapacityCount
	case models.PurchaseConfirmed:
		if departure.ConfirmedCapacity < purchase.CapacityCount {
			return errPurchaseState
		}
		departure.ConfirmedCapacity -= purchase.CapacityCount
	default:
		return errPurchaseState
	}
	return nil
}

func updateDepartureAfterRelease(departure *models.TourPackageDeparture, now time.Time) {
	if departure.Status == "cancelled" || departure.Status == "completed" || departure.Status == "minimum_review" {
		return
	}
	if !departure.BookingClosesAt.After(now) && departure.StartsAt.After(now) && departure.ConfirmedCapacity < departure.MinCapacity {
		markedAt := now.UTC()
		departure.Status = "minimum_review"
		departure.MinimumReviewAt = &markedAt
		return
	}
	if departure.ConfirmedCapacity >= departure.MinCapacity {
		departure.Status = "confirmed"
	} else {
		departure.Status = "open"
	}
}

func (h *PurchaseHandler) CancelMine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input purchaseCancelInput
	if !decodeJSONLimit(w, r, &input, 7<<20) {
		return
	}
	input.Reason = strings.TrimSpace(input.Reason)
	qr, err := validateRefundDestination(&input.refundDestinationInput)
	if input.Version < 1 || len([]rune(input.Reason)) < 5 || len([]rune(input.Reason)) > 1000 || err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PURCHASE_CANCELLATION_INVALID", "Revisa el motivo y los datos para recibir la devolución", nil)
		return
	}
	var purchase models.TourPackagePurchase
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var lookup models.TourPackagePurchase
		if err := tx.Select("id", "departure_id", "package_id").Where("tourist_id = ?", principal.User.ID).First(&lookup, id).Error; err != nil {
			return err
		}
		var departure models.TourPackageDeparture
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&departure, lookup.DepartureID).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("tourist_id = ? AND departure_id = ?", principal.User.ID, departure.ID).First(&purchase, id).Error; err != nil {
			return err
		}
		if purchase.Version != input.Version {
			return errPurchaseConflict
		}
		var item models.TourPackage
		if err := tx.Select("id", "cancellation_allowed", "cancellation_notice_hours").First(&item, lookup.PackageID).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		deadline := departure.StartsAt.Add(-time.Duration(item.CancellationNoticeHours) * time.Hour)
		if !item.CancellationAllowed || !deadline.After(now) || !departure.StartsAt.After(now) {
			return errPurchaseCancellation
		}
		if err := releasePurchaseCapacity(&departure, purchase); err != nil {
			return err
		}
		applyRefundDestination(&purchase, input.refundDestinationInput, qr)
		purchase.CancelledAt = &now
		purchase.CancellationReason = input.Reason
		beginPurchaseRefund(&purchase, "Cancelación solicitada por el turista: "+input.Reason, now)
		purchase.Version++
		updateDepartureAfterRelease(&departure, now)
		departure.Version++
		if err := tx.Save(&departure).Error; err != nil {
			return err
		}
		return tx.Save(&purchase).Error
	})
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	if err := purchaseQuery(h.db).First(&purchase, purchase.ID).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, purchaseView(purchase), "Cancelación registrada; la agencia tiene 72 horas para devolver el pago")
}

func (h *PurchaseHandler) UpdateRefundDestination(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input refundDestinationInput
	if !decodeJSONLimit(w, r, &input, 7<<20) {
		return
	}
	qr, err := validateRefundDestination(&input)
	if input.Version < 1 || err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "REFUND_DESTINATION_INVALID", "Selecciona un QR válido o completa los datos bancarios", nil)
		return
	}
	var purchase models.TourPackagePurchase
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("tourist_id = ?", principal.User.ID).First(&purchase, id).Error; err != nil {
			return err
		}
		if purchase.Version != input.Version {
			return errPurchaseConflict
		}
		if purchase.Status != models.PurchaseRefundPending {
			return errPurchaseState
		}
		applyRefundDestination(&purchase, input, qr)
		purchase.Version++
		return tx.Save(&purchase).Error
	})
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	if err := purchaseQuery(h.db).First(&purchase, purchase.ID).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, purchaseView(purchase), "Datos para el reembolso actualizados")
}

func (h *PurchaseHandler) CompleteRefund(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input refundCompleteInput
	if !decodeJSONLimit(w, r, &input, 7<<20) {
		return
	}
	input.RefundReference = strings.TrimSpace(input.RefundReference)
	proof, err := decodeAttractionImage(input.RefundProof)
	if input.Version < 1 || err != nil || len([]rune(input.RefundReference)) > 160 {
		respond.Error(w, r, http.StatusUnprocessableEntity, "REFUND_PROOF_INVALID", "Sube un comprobante PNG o JPG de hasta 5 MB", nil)
		return
	}
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	var purchase models.TourPackagePurchase
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("agency_id = ?", agency.ID).First(&purchase, id).Error; err != nil {
			return err
		}
		if purchase.Version != input.Version {
			return errPurchaseConflict
		}
		if purchase.Status != models.PurchaseRefundPending || purchase.RefundMethod == "" {
			return errPurchaseState
		}
		now := time.Now().UTC()
		purchase.Status = models.PurchaseRefunded
		purchase.RefundProof = proof
		purchase.RefundReference = input.RefundReference
		purchase.RefundedAt = &now
		purchase.RefundCompletedByUserID = &principal.User.ID
		purchase.Version++
		return tx.Save(&purchase).Error
	})
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	if err := purchaseQuery(h.db).First(&purchase, purchase.ID).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, purchaseView(purchase), "Reembolso completado")
}

func (h *PurchaseHandler) proof(w http.ResponseWriter, r *http.Request, q *gorm.DB) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var purchase models.TourPackagePurchase
	if err := q.Select("id", "payment_proof").First(&purchase, id).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(purchase.PaymentProof)
}

func (h *PurchaseHandler) RefundQRMine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	h.refundImage(w, r, h.db.Where("tourist_id = ?", principal.User.ID), "refund_qr")
}

func (h *PurchaseHandler) RefundQRAgency(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	h.refundImage(w, r, h.db.Where("agency_id = ?", agency.ID), "refund_qr")
}

func (h *PurchaseHandler) RefundProofMine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	h.refundImage(w, r, h.db.Where("tourist_id = ?", principal.User.ID), "refund_proof")
}

func (h *PurchaseHandler) RefundProofAgency(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	h.refundImage(w, r, h.db.Where("agency_id = ?", agency.ID), "refund_proof")
}

func (h *PurchaseHandler) refundImage(w http.ResponseWriter, r *http.Request, q *gorm.DB, column string) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var purchase models.TourPackagePurchase
	if err := q.Select("id", column).First(&purchase, id).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	image := purchase.RefundQR
	if column == "refund_proof" {
		image = purchase.RefundProof
	}
	if len(image) == 0 {
		purchaseError(w, r, gorm.ErrRecordNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(image)
}

func (h *PurchaseHandler) Review(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input purchaseReviewInput
	if !decodeJSON(w, r, &input) {
		return
	}
	input.Decision, input.Reason = strings.TrimSpace(input.Decision), strings.TrimSpace(input.Reason)
	if input.Version < 1 || !slices.Contains([]string{"confirm", "request_correction", "refund"}, input.Decision) || (input.Decision != "confirm" && len([]rune(input.Reason)) < 5) || len([]rune(input.Reason)) > 1000 {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PURCHASE_REVIEW_INVALID", "Revisa la decisión y el motivo", nil)
		return
	}
	var purchase models.TourPackagePurchase
	err := h.db.Transaction(func(tx *gorm.DB) error {
		agency, err := ownAgencyForPurchase(r, tx)
		if err != nil {
			return err
		}
		var purchaseLookup models.TourPackagePurchase
		if err := tx.Select("id", "departure_id").Where("agency_id = ?", agency.ID).First(&purchaseLookup, id).Error; err != nil {
			return err
		}
		var departure models.TourPackageDeparture
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&departure, purchaseLookup.DepartureID).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("agency_id = ? AND departure_id = ?", agency.ID, departure.ID).First(&purchase, id).Error; err != nil {
			return err
		}
		if purchase.Version != input.Version {
			return errPurchaseConflict
		}
		if purchase.Status != models.PurchasePaymentReview {
			return errPurchaseState
		}
		if departure.HeldCapacity < purchase.CapacityCount {
			return errPurchaseState
		}
		now := time.Now().UTC()
		if input.Decision == "confirm" {
			departure.HeldCapacity -= purchase.CapacityCount
			if departure.ConfirmedCapacity+purchase.CapacityCount > departure.MaxCapacity {
				return errPurchaseCapacity
			}
			departure.ConfirmedCapacity += purchase.CapacityCount
			if departure.ConfirmedCapacity >= departure.MinCapacity && slices.Contains([]string{"open", "confirmed", "minimum_review"}, departure.Status) {
				departure.Status = "confirmed"
				if departure.MinimumReviewAt != nil {
					departure.MinimumReviewedAt = &now
					departure.MinimumReviewUserID = &principal.User.ID
				}
			}
			purchase.Status = models.PurchaseConfirmed
			purchase.RejectionReason = ""
		} else if input.Decision == "request_correction" {
			purchase.Status = models.PurchaseCorrectionRequested
			purchase.RejectionReason = input.Reason
		} else {
			departure.HeldCapacity -= purchase.CapacityCount
			beginPurchaseRefund(&purchase, input.Reason, now)
			updateDepartureAfterRelease(&departure, now)
		}
		purchase.AgencyReviewUserID = &principal.User.ID
		purchase.ReviewedAt = &now
		purchase.Version++
		if input.Decision != "request_correction" {
			departure.Version++
			if err := tx.Save(&departure).Error; err != nil {
				return err
			}
		}
		return tx.Save(&purchase).Error
	})
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	if err := purchaseQuery(h.db).First(&purchase, purchase.ID).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, purchaseView(purchase), "Revisión de pago registrada")
}

func ownAgencyForPurchase(r *http.Request, db *gorm.DB) (models.Agency, error) {
	principal, _ := middleware.CurrentPrincipal(r)
	var agency models.Agency
	err := db.Where("manager_id = ?", principal.User.ID).First(&agency).Error
	return agency, err
}

func purchaseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		respond.Error(w, r, http.StatusNotFound, "PURCHASE_NOT_FOUND", "La compra, salida o paquete no existe", nil)
	case errors.Is(err, errPurchaseUnavailable):
		respond.Error(w, r, http.StatusConflict, "DEPARTURE_NOT_BOOKABLE", "La salida ya no está disponible para compra", nil)
	case errors.Is(err, errPurchaseCapacity):
		respond.Error(w, r, http.StatusConflict, "CAPACITY_UNAVAILABLE", "Ya no hay cupos suficientes para este grupo", nil)
	case errors.Is(err, errPurchasePayment):
		respond.Error(w, r, http.StatusUnprocessableEntity, "PAYMENT_METHOD_UNAVAILABLE", "La agencia no admite ese medio de pago", nil)
	case errors.Is(err, errPurchaseConflict):
		respond.Error(w, r, http.StatusConflict, "PURCHASE_CONFLICT", "La compra cambió. Recarga antes de revisarla", nil)
	case errors.Is(err, errPurchaseState):
		respond.Error(w, r, http.StatusConflict, "PURCHASE_STATE_CONFLICT", "La compra ya fue revisada o sus cupos cambiaron", nil)
	case errors.Is(err, errPurchaseCancellation):
		respond.Error(w, r, http.StatusConflict, "PURCHASE_CANCELLATION_UNAVAILABLE", "La política o el plazo de este paquete ya no permiten cancelar", nil)
	case errors.Is(err, errRefundDestination):
		respond.Error(w, r, http.StatusUnprocessableEntity, "REFUND_DESTINATION_INVALID", "Selecciona un QR válido o completa los datos bancarios", nil)
	default:
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "No se pudo completar la operación de compra", nil)
	}
}
