package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

// Amounts describe reviewed purchases, not a bank settlement or an accounting balance.
type departurePayments struct {
	models.TourPackageDeparture
	PackageName          string `json:"package_name"`
	AuthorizedByName     string `json:"authorized_by_name"`
	TotalPurchases       int    `json:"total_purchases"`
	PaymentsPending      int    `json:"payments_pending"`
	CorrectionsPending   int    `json:"corrections_pending"`
	ConfirmedPurchases   int    `json:"confirmed_purchases"`
	RefundsPending       int    `json:"refunds_pending"`
	RefundsOverdue       int    `json:"refunds_overdue"`
	RefundedPurchases    int    `json:"refunded_purchases"`
	ConfirmedAmountCents int64  `json:"confirmed_amount_cents"`
	RefundAmountCents    int64  `json:"refund_amount_cents"`
	ConfirmedTravelers   int    `json:"confirmed_travelers"`
	ConfirmedAdults      int    `json:"confirmed_adults"`
	ConfirmedMinors      int    `json:"confirmed_minors"`
	ConfirmedFreeMinors  int    `json:"confirmed_free_minors"`
	AuthorizationState   string `gorm:"-" json:"authorization_state"`
	CanAuthorize         bool   `gorm:"-" json:"can_authorize"`
	CanCancel            bool   `gorm:"-" json:"can_cancel"`
	AgencyActive         bool   `json:"agency_active"`
}

type paymentTotals struct {
	Departures           int64 `json:"departures"`
	PaymentsPending      int64 `json:"payments_pending"`
	CorrectionsPending   int64 `json:"corrections_pending"`
	RefundsPending       int64 `json:"refunds_pending"`
	RefundsOverdue       int64 `json:"refunds_overdue"`
	ReadyDepartures      int64 `json:"ready_departures"`
	ConfirmedAmountCents int64 `json:"confirmed_amount_cents"`
}

type paymentNavigation struct {
	PaymentsPending      int64 `json:"payments_pending"`
	CorrectionsPending   int64 `json:"corrections_pending"`
	RefundsPending       int64 `json:"refunds_pending"`
	RefundsOverdue       int64 `json:"refunds_overdue"`
	ReadyDepartures      int64 `json:"ready_departures"`
	AuthorizedDepartures int64 `json:"authorized_departures"`
	PendingDepartures    int64 `json:"pending_departures"`
	RefundDepartures     int64 `json:"refund_departures"`
	HistoryDepartures    int64 `json:"history_departures"`
}

const validAuthorizationSQL = "d.autorizado_en IS NOT NULL AND d.autorizacion_revocada_en IS NULL"
const readyDepartureSQL = "a.estado = 'active' AND d.salida_en > ? AND d.estado IN ('open','confirmed','closed') AND d.cupos_confirmados >= d.cupo_minimo AND (d.autorizado_en IS NULL OR d.autorizacion_revocada_en IS NOT NULL)"

const authorizedDepartureSQL = "a.estado='active' AND d.salida_en > ? AND d.estado IN ('open','confirmed','closed') AND d.cupos_confirmados >= d.cupo_minimo AND (" + validAuthorizationSQL + ")"
const pendingDepartureSQL = "d.salida_en > ? AND d.estado NOT IN ('cancelled','completed') AND NOT (a.estado='active' AND d.estado IN ('open','confirmed','closed') AND d.cupos_confirmados >= d.cupo_minimo AND (" + validAuthorizationSQL + "))"
const refundDepartureSQL = "COALESCE(pagos.refunds_pending,0)+COALESCE(pagos.refunded_purchases,0)>0"
const historyDepartureSQL = "d.salida_en <= ? OR d.estado IN ('cancelled','completed')"

func paymentDepartureQuery(db *gorm.DB, agencyID uint64) *gorm.DB {
	return db.Table("salidas_paquete d").
		Joins("JOIN paquetes_turisticos p ON p.id = d.paquete_id").
		Joins("JOIN agencias a ON a.id = p.agencia_id").
		Joins("LEFT JOIN usuarios u ON u.id = d.usuario_autorizacion_id").
		Joins(`LEFT JOIN (
   SELECT salida_id,
    COUNT(*) AS total_purchases,
    COUNT(*) FILTER (WHERE estado='payment_review') AS payments_pending,
    COUNT(*) FILTER (WHERE estado='correction_requested') AS corrections_pending,
    COUNT(*) FILTER (WHERE estado='confirmed') AS confirmed_purchases,
    COUNT(*) FILTER (WHERE estado='refund_pending') AS refunds_pending,
    COUNT(*) FILTER (WHERE estado='refund_pending' AND reembolso_vence_en <= NOW()) AS refunds_overdue,
    COUNT(*) FILTER (WHERE estado='refunded') AS refunded_purchases,
    COALESCE(SUM(total_centavos) FILTER (WHERE estado='confirmed'),0) AS confirmed_amount_cents,
    COALESCE(SUM(total_centavos) FILTER (WHERE estado='refund_pending'),0) AS refund_amount_cents,
    COALESCE(SUM(adultos_nacionales+adultos_extranjeros+jsonb_array_length(menores)) FILTER (WHERE estado='confirmed'),0) AS confirmed_travelers,
    COALESCE(SUM(adultos_nacionales+adultos_extranjeros) FILTER (WHERE estado='confirmed'),0) AS confirmed_adults,
    COALESCE(SUM(jsonb_array_length(menores)) FILTER (WHERE estado='confirmed'),0) AS confirmed_minors,
    COALESCE(SUM(cantidad_menores_gratis) FILTER (WHERE estado='confirmed'),0) AS confirmed_free_minors
   FROM compras_paquete WHERE agencia_id = ? GROUP BY salida_id
  ) pagos ON pagos.salida_id=d.id`, agencyID).
		Where("p.agencia_id = ?", agencyID)
}

const paymentFields = `d.*, p.nombre AS package_name,
 COALESCE(u.nombres || ' ' || u.apellidos,'') AS authorized_by_name, (a.estado='active') AS agency_active,
 COALESCE(pagos.total_purchases,0) AS total_purchases,
 COALESCE(pagos.payments_pending,0) AS payments_pending,
 COALESCE(pagos.corrections_pending,0) AS corrections_pending,
 COALESCE(pagos.confirmed_purchases,0) AS confirmed_purchases,
 COALESCE(pagos.refunds_pending,0) AS refunds_pending,
 COALESCE(pagos.refunds_overdue,0) AS refunds_overdue,
 COALESCE(pagos.refunded_purchases,0) AS refunded_purchases,
 COALESCE(pagos.confirmed_amount_cents,0) AS confirmed_amount_cents,
 COALESCE(pagos.refund_amount_cents,0) AS refund_amount_cents,
 COALESCE(pagos.confirmed_travelers,0) AS confirmed_travelers,
 COALESCE(pagos.confirmed_adults,0) AS confirmed_adults,
 COALESCE(pagos.confirmed_minors,0) AS confirmed_minors,
 COALESCE(pagos.confirmed_free_minors,0) AS confirmed_free_minors`

func authorizationState(d models.TourPackageDeparture, active bool, now time.Time) string {
	if d.Status == "cancelled" {
		return "cancelled"
	}
	if d.Status == "completed" || !d.StartsAt.After(now) {
		return "past"
	}
	if !active || d.Status == "draft" {
		return "unavailable"
	}
	if d.ConfirmedCapacity < d.MinCapacity {
		if d.Status == "minimum_review" || !d.BookingClosesAt.After(now) {
			return "minimum_review"
		}
		return "minimum_pending"
	}
	if d.AuthorizedAt != nil && d.AuthorizationRevokedAt == nil {
		return "authorized"
	}
	if slices.Contains([]string{"open", "confirmed", "closed"}, d.Status) {
		return "ready"
	}
	return "unavailable"
}

func decorateDeparture(d *departurePayments, now time.Time) {
	d.AuthorizationState = authorizationState(d.TourPackageDeparture, d.AgencyActive, now)
	d.CanAuthorize = d.AuthorizationState == "ready"
	d.CanCancel = d.AgencyActive && d.StartsAt.After(now) && d.Status != "cancelled" && d.Status != "completed"
}

func paymentFilterError(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, r, 422, "PAYMENT_FILTER_INVALID", "Revisa el paquete, las fechas y el estado seleccionados", nil)
}

func (h *PurchaseHandler) PaymentDepartures(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	now := time.Now().UTC()
	q := paymentDepartureQuery(h.db, agency.ID)
	if value := r.URL.Query().Get("package_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil || id == 0 {
			paymentFilterError(w, r)
			return
		}
		q = q.Where("d.paquete_id = ?", id)
	}
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		q = q.Where("LOWER(p.nombre) LIKE ? ESCAPE '!'", "%"+escapeLike(strings.ToLower(search))+"%")
	}
	var from, to time.Time
	for _, key := range []string{"from", "to"} {
		if value := r.URL.Query().Get(key); value != "" {
			day, err := time.ParseInLocation("2006-01-02", value, boliviaTime)
			if err != nil {
				paymentFilterError(w, r)
				return
			}
			if key == "from" {
				from = day
				q = q.Where("d.salida_en >= ?", day)
			} else {
				to = day
				q = q.Where("d.salida_en < ?", day.AddDate(0, 0, 1))
			}
		}
	}
	if !from.IsZero() && !to.IsZero() && from.After(to) {
		paymentFilterError(w, r)
		return
	}
	// Navigation badges and payment alerts retain unresolved work outside the
	// main authorized-departures view. Package/search/date filters still apply.
	var summary paymentNavigation
	err = q.Session(&gorm.Session{}).Select(`
 COALESCE(SUM(pagos.payments_pending),0) AS payments_pending,
 COALESCE(SUM(pagos.corrections_pending),0) AS corrections_pending,
 COALESCE(SUM(pagos.refunds_pending),0) AS refunds_pending,
 COALESCE(SUM(pagos.refunds_overdue),0) AS refunds_overdue,
 COUNT(*) FILTER (WHERE `+readyDepartureSQL+`) AS ready_departures,
 COUNT(*) FILTER (WHERE `+authorizedDepartureSQL+`) AS authorized_departures,
 COUNT(*) FILTER (WHERE `+pendingDepartureSQL+`) AS pending_departures,
 COUNT(*) FILTER (WHERE `+refundDepartureSQL+`) AS refund_departures,
 COUNT(*) FILTER (WHERE `+historyDepartureSQL+`) AS history_departures`, now, now, now, now).Scan(&summary).Error
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	state := r.URL.Query().Get("state")
	switch state {
	case "", "all":
	case "ready":
		q = q.Where(readyDepartureSQL, now)
	case "authorized":
		q = q.Where(authorizedDepartureSQL, now)
	case "pending":
		q = q.Where(pendingDepartureSQL, now)
	case "refunds":
		q = q.Where(refundDepartureSQL)
	case "history":
		q = q.Where(historyDepartureSQL, now)
	case "minimum_pending":
		q = q.Where("d.cupos_confirmados < d.cupo_minimo AND d.compras_cierran_en > ? AND d.estado IN ('open','confirmed')", now)
	case "minimum_review":
		q = q.Where("d.salida_en > ? AND d.cupos_confirmados < d.cupo_minimo AND d.compras_cierran_en <= ? AND d.estado NOT IN ('cancelled','completed','draft')", now, now)
	case "cancelled":
		q = q.Where("d.estado = 'cancelled'")
	case "past":
		q = q.Where("d.salida_en <= ? OR d.estado='completed'", now)
	case "attention":
		q = q.Where("COALESCE(pagos.payments_pending,0)+COALESCE(pagos.corrections_pending,0)+COALESCE(pagos.refunds_pending,0)>0 OR ("+readyDepartureSQL+") OR (d.estado='minimum_review' AND d.salida_en > ?)", now, now)
	default:
		paymentFilterError(w, r)
		return
	}
	// Every total is computed over the full filtered set, before pagination.
	var totals paymentTotals
	err = q.Session(&gorm.Session{}).Select(`COUNT(*) AS departures,
 COALESCE(SUM(pagos.payments_pending),0) AS payments_pending,
 COALESCE(SUM(pagos.corrections_pending),0) AS corrections_pending,
 COALESCE(SUM(pagos.refunds_pending),0) AS refunds_pending,
 COALESCE(SUM(pagos.refunds_overdue),0) AS refunds_overdue,
 COUNT(*) FILTER (WHERE `+readyDepartureSQL+`) AS ready_departures,
 COALESCE(SUM(pagos.confirmed_amount_cents),0) AS confirmed_amount_cents`, now).Scan(&totals).Error
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	page, limit := queryInt(r, "page", 1, 1, 100000), queryInt(r, "limit", 10, 1, 50)
	rows := []departurePayments{}
	err = q.Session(&gorm.Session{}).Select(paymentFields).Clauses(clause.OrderBy{Expression: clause.Expr{SQL: `CASE WHEN COALESCE(pagos.payments_pending,0)+COALESCE(pagos.corrections_pending,0)+COALESCE(pagos.refunds_pending,0)>0 OR (` + readyDepartureSQL + `) OR (d.estado='minimum_review' AND d.salida_en > ?) THEN 0 WHEN d.salida_en > ? THEN 1 ELSE 2 END, d.salida_en, d.id`, Vars: []any{now, now, now}, WithoutParentheses: true}}).Limit(limit).Offset((page - 1) * limit).Scan(&rows).Error
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	for i := range rows {
		decorateDeparture(&rows[i], now)
	}
	packages := []struct {
		ID   uint64 `json:"id"`
		Name string `gorm:"column:nombre" json:"name"`
	}{}
	if err := h.db.Model(&models.TourPackage{}).Select("id,nombre").Where("agencia_id = ?", agency.ID).Order("nombre,id").Scan(&packages).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	respond.JSON(w, 200, map[string]any{"departures": rows, "packages": packages, "totals": totals, "summary": summary, "pagination": map[string]any{"page": page, "limit": limit, "total": totals.Departures}}, "")
}

func (h *PurchaseHandler) PaymentDeparture(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	id, ok := departurePathID(w, r)
	if !ok {
		return
	}
	var row departurePayments
	if err := paymentDepartureQuery(h.db, agency.ID).Select(paymentFields).Where("d.id = ?", id).Take(&row).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	decorateDeparture(&row, time.Now())
	respond.JSON(w, 200, row, "")
}

var errDepartureNotReady = errors.New("departure not ready for authorization")

func (h *PackageHandler) AuthorizeDeparture(w http.ResponseWriter, r *http.Request) {
	id, ok := departurePathID(w, r)
	if !ok {
		return
	}
	var input minimumRefundInput
	if !decodeJSONLimit(w, r, &input, 8<<10) {
		return
	}
	if input.Version < 1 {
		respond.Error(w, r, 422, "DEPARTURE_VALIDATION", "Recarga la salida antes de confirmarla", nil)
		return
	}
	var departure models.TourPackageDeparture
	err := h.db.Transaction(func(tx *gorm.DB) error {
		agency, err := h.ownAgency(r, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		if err != nil {
			return err
		}
		if agency.Status != models.StatusActive {
			return errPackageAgencyInactive
		}
		packages := tx.Model(&models.TourPackage{}).Select("id").Where("agencia_id = ?", agency.ID)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("paquete_id IN (?)", packages).First(&departure, id).Error; err != nil {
			return err
		}
		if departure.Version != input.Version {
			return errDepartureConflict
		}
		now := time.Now().UTC()
		if authorizationState(departure, true, now) != "ready" {
			return errDepartureNotReady
		}
		principal, _ := middleware.CurrentPrincipal(r)
		departure.AuthorizedAt = &now
		departure.AuthorizedByUserID = &principal.User.ID
		departure.AuthorizationRevokedAt = nil
		departure.AuthorizationRevocationReason = ""
		departure.Version++
		return tx.Save(&departure).Error
	})
	if errors.Is(err, errDepartureNotReady) {
		respond.Error(w, r, 422, "DEPARTURE_AUTHORIZATION_UNAVAILABLE", "La salida requiere el mínimo de cupos confirmados, una fecha futura y no estar autorizada ni cancelada. Recarga su estado", nil)
		return
	}
	if err != nil {
		departureError(w, r, err)
		return
	}
	respond.JSON(w, 200, departure, "Salida autorizada por el encargado")
}

type departureRosterEntry struct {
	Reference string `gorm:"column:referencia"`
	Name      string `gorm:"column:nombre_turista"`
	Phone     string `gorm:"column:telefono_turista"`
	Email     string `gorm:"column:correo_turista"`
	Adults    int
	Minors    int
	Travelers int
	Capacity  int `gorm:"column:cantidad_cupos"`
}

// The printable list deliberately includes every confirmed purchase, not just
// the currently displayed page. It never includes identity documents or proofs.
func (h *PurchaseHandler) DepartureRoster(w http.ResponseWriter, r *http.Request) {
	agency, err := ownAgencyForPurchase(r, h.db)
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	id, ok := departurePathID(w, r)
	if !ok {
		return
	}
	var departure departurePayments
	if err := paymentDepartureQuery(h.db, agency.ID).Select(paymentFields).Where("d.id = ?", id).Take(&departure).Error; err != nil {
		purchaseError(w, r, err)
		return
	}
	rows := []departureRosterEntry{}
	err = h.db.Model(&models.TourPackagePurchase{}).
		Select("referencia,nombre_turista,telefono_turista,correo_turista,adultos_nacionales+adultos_extranjeros AS adults,jsonb_array_length(menores) AS minors,adultos_nacionales+adultos_extranjeros+jsonb_array_length(menores) AS travelers,cantidad_cupos").
		Where("agencia_id=? AND salida_id=? AND estado=?", agency.ID, id, models.PurchaseConfirmed).
		Order("nombre_turista,id").Scan(&rows).Error
	if err != nil {
		purchaseError(w, r, err)
		return
	}
	decorateDeparture(&departure, time.Now())
	// Printed totals come from these exact rows, not a separately read dashboard.
	travelers, capacity := 0, 0
	for _, row := range rows {
		travelers += row.Travelers
		capacity += row.Capacity
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	// html/template escapes all names, references and contact data.
	_ = rosterTemplate.Execute(w, struct {
		Agency, Package, Date, Generated, State string
		Rows                                    []departureRosterEntry
		Travelers, Capacity                     int
	}{
		Agency: agency.Name, Package: departure.PackageName, Date: departure.StartsAt.In(boliviaTime).Format("02/01/2006 15:04"),
		Generated: time.Now().In(boliviaTime).Format("02/01/2006 15:04"), State: map[string]string{"authorized": "Salida autorizada", "ready": "Pendiente de autorizacion", "cancelled": "Salida cancelada", "past": "Salida pasada", "minimum_review": "Revision por minimo", "minimum_pending": "Minimo pendiente", "unavailable": "No habilitada"}[departure.AuthorizationState],
		Rows: rows, Travelers: travelers, Capacity: capacity,
	})
}

var rosterTemplate = template.Must(template.New("roster").Parse(`<!doctype html>
<html lang="es"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Lista de turistas — {{.Package}}</title>
<style>body{font:14px/1.5 system-ui,sans-serif;color:#111;margin:32px}h1{font-size:26px;margin:6px 0}p{margin:6px 0}table{border-collapse:collapse;width:100%;margin:24px 0}th,td{text-align:left;padding:10px 8px;border-bottom:1px solid #ccc;vertical-align:top;overflow-wrap:anywhere}th{background:#f0f2f3}small{display:block}.notice{padding:10px;border:1px solid #ddd;margin-top:14px}@page{size:A4 landscape;margin:14mm}@media print{body{margin:0;font-size:10pt}thead{display:table-header-group}tr{break-inside:avoid}.screen-only{display:none}}</style>
<header><p>{{.Agency}}</p><h1>{{.Package}}</h1><p>Salida: {{.Date}} · Hora de Bolivia</p><p>{{.State}} · {{.Travelers}} {{if eq .Travelers 1}}turista{{else}}turistas{{end}} · {{.Capacity}} cupos confirmados</p><p class="screen-only">Utiliza la opción Imprimir de tu navegador para imprimir o guardar como PDF.</p></header>
<p class="notice">Solo compras con pago confirmado. Cada fila corresponde al comprador responsable de un grupo; no es una identificación individual de todos los acompañantes.</p>
<table><thead><tr><th>Referencia</th><th>Comprador responsable</th><th>Contacto</th><th>Adultos</th><th>Menores</th><th>Turistas</th><th>Cupos</th></tr></thead><tbody>{{range .Rows}}<tr><td>{{.Reference}}</td><td>{{.Name}}</td><td>{{.Phone}}<small>{{.Email}}</small></td><td>{{.Adults}}</td><td>{{.Minors}}</td><td>{{.Travelers}}</td><td>{{.Capacity}}</td></tr>{{else}}<tr><td colspan="7">No hay compras con pago confirmado.</td></tr>{{end}}</tbody></table>
<footer>Generado: {{.Generated}} (Bolivia). Uso interno de la agencia. Los menores exentos de pago no consumen cupo según la política vigente.</footer></html>`))
