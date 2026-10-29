package handlers

import (
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/receipts"
	"starter-backend/internal/respond"
	"time"
)

func (h *PurchaseHandler) ReceiptMine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	h.receipt(w, r, h.db.Where("turista_id = ?", principal.User.ID))
}
func (h *PurchaseHandler) ReceiptAgency(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	h.receipt(w, r, h.db.Where("agencia_id = ?", agency.ID))
}
func (h *PurchaseHandler) receipt(w http.ResponseWriter, r *http.Request, q *gorm.DB) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var p models.TourPackagePurchase
	// Binary proofs and private refund destinations never enter the renderer.
	if err := q.Omit("comprobante_pago", "qr_reembolso", "comprobante_reembolso", "nombre_banco_reembolso", "titular_cuenta_reembolso", "numero_cuenta_reembolso").Preload("Departure").Preload("Agency").First(&p, id).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	if !receipts.Available(p.Status) {
		respond.Error(w, r, 409, "PURCHASE_RECEIPT_UNAVAILABLE", "El comprobante PDF estará disponible cuando la agencia confirme el pago", nil)
		return
	}
	if p.Departure == nil || p.Agency == nil || p.ReceiptSnapshot.Origin == "" {
		respond.Error(w, r, 500, "PURCHASE_RECEIPT_ERROR", "No se pudo preparar el comprobante PDF", nil)
		return
	}
	data, err := receipts.Purchase(p, *p.Departure, p.Agency.Status == models.StatusActive, time.Now().UTC())
	if err != nil {
		respond.Error(w, r, 500, "PURCHASE_RECEIPT_ERROR", "No se pudo generar el comprobante PDF", nil)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	disposition := "inline"
	if r.URL.Query().Get("download") == "1" {
		disposition = "attachment"
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`%s; filename="andaria-compra-%d.pdf"`, disposition, p.ID))
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	_, _ = w.Write(data)
}
