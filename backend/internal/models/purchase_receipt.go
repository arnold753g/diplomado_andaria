package models

import "time"

// PurchaseReceiptSnapshot is captured by PostgreSQL when the purchase is created.
// It never accepts client data and is deliberately excluded from public JSON.
type PurchaseReceiptSnapshot struct {
	Origin                  string    `json:"origen"`
	CapturedAt              time.Time `json:"registrado_en"`
	AgencyName              string    `json:"agencia_nombre"`
	AgencyAddress           string    `json:"agencia_direccion"`
	AgencyPhone             string    `json:"agencia_telefono"`
	AgencyEmail             string    `json:"agencia_correo"`
	PackageName             string    `json:"paquete_nombre"`
	DurationDays            int       `json:"duracion_dias"`
	DurationNights          int       `json:"duracion_noches"`
	Bring                   []string  `json:"llevar"`
	CancellationAllowed     bool      `json:"permite_cancelacion"`
	CancellationNoticeHours int       `json:"horas_aviso_cancelacion"`
	DepartureAt             time.Time `json:"salida_en"`
	MeetingAt               time.Time `json:"encuentro_en"`
	MeetingPoint            string    `json:"punto_encuentro"`
}
