package models

import "time"

const (
	PurchasePaymentReview       = "payment_review"
	PurchaseCorrectionRequested = "correction_requested"
	PurchaseConfirmed           = "confirmed"
	PurchasePaymentRejected     = "payment_rejected"
	PurchaseCancelled           = "cancelled"
	PurchaseRefundPending       = "refund_pending"
	PurchaseRefunded            = "refunded"
)

type PurchaseMinor struct {
	Age       int  `json:"age"`
	IsForeign bool `json:"is_foreign"`
	Pays      bool `json:"pays"`
}

type TourPackagePurchase struct {
	ID                      uint64                `gorm:"primaryKey" json:"id"`
	Reference               string                `json:"reference"`
	TouristID               uint64                `json:"tourist_id"`
	AgencyID                uint64                `json:"agency_id"`
	PackageID               uint64                `json:"package_id"`
	DepartureID             uint64                `json:"departure_id"`
	Status                  string                `json:"status"`
	PaymentMethod           string                `json:"payment_method"`
	NationalAdults          int                   `json:"national_adults"`
	ForeignAdults           int                   `json:"foreign_adults"`
	Minors                  []PurchaseMinor       `gorm:"serializer:json" json:"minors"`
	FreeMinorCount          int                   `json:"free_minor_count"`
	PayingMinorCount        int                   `json:"paying_minor_count"`
	CapacityCount           int                   `json:"capacity_count"`
	NationalUnitPriceCents  int                   `json:"national_unit_price_cents"`
	ForeignSurchargeCents   int                   `json:"foreign_surcharge_cents"`
	TotalCents              int                   `json:"total_cents"`
	TouristName             string                `json:"tourist_name"`
	TouristEmail            string                `json:"tourist_email"`
	TouristPhone            string                `json:"tourist_phone"`
	TouristDocument         string                `json:"tourist_document"`
	TouristNationality      string                `json:"tourist_nationality"`
	PaymentProof            []byte                `json:"-"`
	AgencyReviewUserID      *uint64               `json:"agency_review_user_id,omitempty"`
	ReviewedAt              *time.Time            `json:"reviewed_at,omitempty"`
	RejectionReason         string                `json:"rejection_reason"`
	CancelledAt             *time.Time            `json:"cancelled_at,omitempty"`
	CancellationReason      string                `json:"cancellation_reason"`
	RefundReason            string                `json:"refund_reason"`
	RefundRequestedAt       *time.Time            `json:"refund_requested_at,omitempty"`
	RefundDueAt             *time.Time            `json:"refund_due_at,omitempty"`
	RefundMethod            string                `json:"refund_method,omitempty"`
	RefundQR                []byte                `json:"-"`
	RefundBankName          string                `json:"refund_bank_name,omitempty"`
	RefundAccountHolder     string                `json:"refund_account_holder,omitempty"`
	RefundAccountNumber     string                `json:"refund_account_number,omitempty"`
	RefundProof             []byte                `json:"-"`
	RefundReference         string                `json:"refund_reference,omitempty"`
	RefundedAt              *time.Time            `json:"refunded_at,omitempty"`
	RefundCompletedByUserID *uint64               `json:"refund_completed_by_user_id,omitempty"`
	Version                 int                   `json:"version"`
	Package                 *TourPackage          `gorm:"foreignKey:PackageID" json:"-"`
	Departure               *TourPackageDeparture `gorm:"foreignKey:DepartureID" json:"-"`
	Agency                  *Agency               `gorm:"foreignKey:AgencyID" json:"-"`
	CreatedAt               time.Time             `json:"created_at"`
	UpdatedAt               time.Time             `json:"updated_at"`
}

func (TourPackagePurchase) TableName() string { return "tour_package_purchases" }
