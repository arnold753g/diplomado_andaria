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
	ID                      uint64                  `gorm:"column:id;primaryKey" json:"id"`
	Reference               string                  `gorm:"column:referencia" json:"reference"`
	ReceiptSnapshot         PurchaseReceiptSnapshot `gorm:"column:datos_comprobante;serializer:json" json:"-"`
	PaymentConfirmedAt      *time.Time              `gorm:"column:pago_confirmado_en" json:"payment_confirmed_at,omitempty"`
	TouristID               uint64                  `gorm:"column:turista_id" json:"tourist_id"`
	AgencyID                uint64                  `gorm:"column:agencia_id" json:"agency_id"`
	PackageID               uint64                  `gorm:"column:paquete_id" json:"package_id"`
	DepartureID             uint64                  `gorm:"column:salida_id" json:"departure_id"`
	Status                  string                  `gorm:"column:estado" json:"status"`
	PaymentMethod           string                  `gorm:"column:metodo_pago" json:"payment_method"`
	NationalAdults          int                     `gorm:"column:adultos_nacionales" json:"national_adults"`
	ForeignAdults           int                     `gorm:"column:adultos_extranjeros" json:"foreign_adults"`
	Minors                  []PurchaseMinor         `gorm:"column:menores;serializer:json" json:"minors"`
	FreeMinorCount          int                     `gorm:"column:cantidad_menores_gratis" json:"free_minor_count"`
	PayingMinorCount        int                     `gorm:"column:cantidad_menores_pago" json:"paying_minor_count"`
	CapacityCount           int                     `gorm:"column:cantidad_cupos" json:"capacity_count"`
	NationalUnitPriceCents  int                     `gorm:"column:precio_unitario_nacional_centavos" json:"national_unit_price_cents"`
	ForeignSurchargeCents   int                     `gorm:"column:recargo_extranjero_centavos" json:"foreign_surcharge_cents"`
	TotalCents              int                     `gorm:"column:total_centavos" json:"total_cents"`
	TouristName             string                  `gorm:"column:nombre_turista" json:"tourist_name"`
	TouristEmail            string                  `gorm:"column:correo_turista" json:"tourist_email"`
	TouristPhone            string                  `gorm:"column:telefono_turista" json:"tourist_phone"`
	TouristDocument         string                  `gorm:"column:documento_turista" json:"tourist_document"`
	TouristNationality      string                  `gorm:"column:nacionalidad_turista" json:"tourist_nationality"`
	PaymentProof            []byte                  `gorm:"column:comprobante_pago" json:"-"`
	AgencyReviewUserID      *uint64                 `gorm:"column:usuario_revision_agencia_id" json:"agency_review_user_id,omitempty"`
	ReviewedAt              *time.Time              `gorm:"column:revisado_en" json:"reviewed_at,omitempty"`
	RejectionReason         string                  `gorm:"column:motivo_rechazo" json:"rejection_reason"`
	CancelledAt             *time.Time              `gorm:"column:cancelado_en" json:"cancelled_at,omitempty"`
	CancellationReason      string                  `gorm:"column:motivo_cancelacion" json:"cancellation_reason"`
	RefundReason            string                  `gorm:"column:motivo_reembolso" json:"refund_reason"`
	RefundRequestedAt       *time.Time              `gorm:"column:reembolso_solicitado_en" json:"refund_requested_at,omitempty"`
	RefundDueAt             *time.Time              `gorm:"column:reembolso_vence_en" json:"refund_due_at,omitempty"`
	RefundMethod            string                  `gorm:"column:metodo_reembolso" json:"refund_method,omitempty"`
	RefundQR                []byte                  `gorm:"column:qr_reembolso" json:"-"`
	RefundBankName          string                  `gorm:"column:nombre_banco_reembolso" json:"refund_bank_name,omitempty"`
	RefundAccountHolder     string                  `gorm:"column:titular_cuenta_reembolso" json:"refund_account_holder,omitempty"`
	RefundAccountNumber     string                  `gorm:"column:numero_cuenta_reembolso" json:"refund_account_number,omitempty"`
	RefundProof             []byte                  `gorm:"column:comprobante_reembolso" json:"-"`
	RefundReference         string                  `gorm:"column:referencia_reembolso" json:"refund_reference,omitempty"`
	RefundedAt              *time.Time              `gorm:"column:reembolsado_en" json:"refunded_at,omitempty"`
	RefundCompletedByUserID *uint64                 `gorm:"column:usuario_reembolso_id" json:"refund_completed_by_user_id,omitempty"`
	Version                 int                     `gorm:"column:version" json:"version"`
	Package                 *TourPackage            `gorm:"foreignKey:PackageID" json:"-"`
	Departure               *TourPackageDeparture   `gorm:"foreignKey:DepartureID" json:"-"`
	Agency                  *Agency                 `gorm:"foreignKey:AgencyID" json:"-"`
	CreatedAt               time.Time               `gorm:"column:creado_en" json:"created_at"`
	UpdatedAt               time.Time               `gorm:"column:actualizado_en" json:"updated_at"`
}

func (TourPackagePurchase) TableName() string { return "compras_paquete" }
