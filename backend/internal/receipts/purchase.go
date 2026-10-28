package receipts

import (
	"bytes"
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/signintech/gopdf"
	"starter-backend/internal/models"
)

// Bundled assets work in the minimal Docker image without network/font services.
//
//go:embed assets/Outfit-Regular.ttf
var font []byte

//go:embed assets/Outfit-SemiBold.ttf
var strongFont []byte

//go:embed assets/andaria-symbol.png
var logo []byte

var bolivia = time.FixedZone("America/La_Paz", -4*60*60)

func Available(status string) bool {
	return status == models.PurchaseConfirmed || status == models.PurchaseRefundPending || status == models.PurchaseRefunded
}

func money(cents int) string { return fmt.Sprintf("Bs %d,%02d", cents/100, cents%100) }
func date(t time.Time) string {
	if t.IsZero() {
		return "Sin registro"
	}
	return t.In(bolivia).Format("02/01/2006 15:04")
}

type row struct {
	category    string
	count, unit int
}

func rows(p models.TourPackagePurchase) []row {
	national, foreign, free := 0, 0, 0
	for _, m := range p.Minors {
		if !m.Pays {
			free++
		} else if m.IsForeign {
			foreign++
		} else {
			national++
		}
	}
	return []row{
		{"Adultos nacionales", p.NationalAdults, p.NationalUnitPriceCents},
		{"Adultos extranjeros", p.ForeignAdults, p.NationalUnitPriceCents + p.ForeignSurchargeCents},
		{"Menores nacionales", national, p.NationalUnitPriceCents},
		{"Menores extranjeros", foreign, p.NationalUnitPriceCents + p.ForeignSurchargeCents},
		{"Menores exentos", free, 0},
	}
}

type document struct {
	pdf       gopdf.GoPdf
	y         float64
	err       error
	reference string
}

func (d *document) text(x, y, width float64, value string, size int) {
	if d.err != nil {
		return
	}
	style := ""
	if size >= 11 {
		style = "B"
	}
	if d.err = d.pdf.SetFont("Outfit", style, size); d.err != nil {
		return
	}
	d.pdf.SetXY(x, y)
	d.err = d.pdf.Cell(&gopdf.Rect{W: width, H: 16}, value)
}
func (d *document) page() {
	d.pdf.AddPage()
	d.pdf.SetTextColor(86, 104, 57)
	d.text(38, 27, 350, "ANDARIA  /  COMPROBANTE DE COMPRA", 11)
	d.pdf.SetTextColor(90, 95, 100)
	d.text(38, 810, 450, fmt.Sprintf("%s  |  Página %d  |  Horarios de Bolivia (UTC-4)", d.reference, d.pdf.GetNumberOfPages()), 8)
	d.pdf.SetTextColor(23, 30, 27)
	d.y = 62
}
func (d *document) space(height float64) {
	if d.y+height > 785 {
		d.page()
	}
}
func (d *document) section(title string) {
	d.space(52)
	d.y += 4
	d.pdf.SetFillColor(239, 243, 233)
	d.pdf.RectFromUpperLeftWithStyle(38, d.y, 519, 24, "F")
	d.pdf.SetTextColor(66, 83, 42)
	d.text(46, d.y+3, 503, title, 11)
	d.pdf.SetTextColor(23, 30, 27)
	d.y += 26
}
func clean(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' {
			return ' '
		}
		return r
	}, value)
}
func (d *document) paragraph(value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if d.err != nil {
		return
	}
	d.err = d.pdf.SetFont("Outfit", "", 10)
	// Rune wrapping also handles very long references, emails and unbroken text.
	lines := []string{}
	for _, paragraph := range strings.Split(clean(value), "\n") {
		line := ""
		for _, word := range strings.Fields(paragraph) {
			candidate := strings.TrimSpace(line + " " + word)
			width, err := d.pdf.MeasureTextWidth(candidate)
			if err != nil {
				d.err = err
				return
			}
			if width <= 515 {
				line = candidate
				continue
			}
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			for _, r := range word {
				width, err = d.pdf.MeasureTextWidth(line + string(r))
				if err != nil {
					d.err = err
					return
				}
				if width > 515 {
					lines = append(lines, line)
					line = ""
				}
				line += string(r)
			}
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	// Keep short paragraphs together, including the final document notice.
	if len(lines) <= 8 {
		d.space(float64(len(lines)*13 + 1))
	}
	for _, line := range lines {
		d.line(line)
	}
	d.y += 1
}
func (d *document) line(value string)        { d.space(15); d.text(40, d.y, 515, value, 10); d.y += 13 }
func (d *document) pair(label, value string) { d.paragraph(label + ": " + value) }

// Purchase uses the immutable commercial snapshot and original integer prices.
// Current financial and operational states are explicitly dated on every copy.
func Purchase(p models.TourPackagePurchase, departure models.TourPackageDeparture, agencyActive bool, now time.Time) ([]byte, error) {
	d := document{reference: p.Reference}
	d.pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := d.pdf.AddTTFFontData("Outfit", font); err != nil {
		return nil, err
	}
	if err := d.pdf.AddTTFFontDataWithOption("Outfit", strongFont, gopdf.TtfOption{Style: gopdf.Bold}); err != nil {
		return nil, err
	}
	d.page()
	holder, err := gopdf.ImageHolderByBytes(logo)
	if err != nil {
		return nil, err
	}
	if err = d.pdf.ImageByHolder(holder, 499, 60, &gopdf.Rect{W: 48, H: 48}); err != nil {
		return nil, err
	}
	s := p.ReceiptSnapshot
	d.text(40, d.y, 450, "Constancia de compra", 22)
	d.y += 33
	d.pair("Referencia", p.Reference)
	d.pair("Compra registrada", date(p.CreatedAt))
	d.pair("Documento generado", date(now))
	d.section("AGENCIA Y COMPRADOR RESPONSABLE")
	d.pair("Agencia", s.AgencyName)
	if s.AgencyAddress != "" {
		d.pair("Dirección", s.AgencyAddress)
	}
	d.paragraph(strings.Join(nonempty(s.AgencyPhone, s.AgencyEmail), "  |  "))
	d.pair("Turista", p.TouristName)
	if p.TouristDocument != "" {
		d.pair("Documento", p.TouristDocument)
	}
	d.paragraph(strings.Join(nonempty(p.TouristEmail, p.TouristPhone), "  |  "))
	d.section("SALIDA CONTRATADA")
	d.pair("Paquete", s.PackageName)
	d.pair("Salida", date(s.DepartureAt))
	d.pair("Encuentro", date(s.MeetingAt))
	if s.MeetingPoint != "" {
		d.pair("Punto de encuentro", s.MeetingPoint)
	}
	d.pair("Duración", fmt.Sprintf("%d días / %d noches", s.DurationDays, s.DurationNights))
	d.section("DETALLE DE TARIFAS ORIGINALES")
	drawHeader := func() {
		d.space(24)
		d.text(40, d.y, 250, "Categoría", 10)
		d.text(302, d.y, 45, "Cant.", 10)
		d.text(353, d.y, 90, "Precio unitario", 10)
		d.text(463, d.y, 92, "Subtotal", 10)
		d.y += 22
	}
	drawHeader()
	for _, r := range rows(p) {
		if r.count == 0 {
			continue
		}
		if d.y+20 > 785 {
			d.page()
			drawHeader()
		}
		d.text(40, d.y, 250, r.category, 10)
		d.text(302, d.y, 45, fmt.Sprint(r.count), 10)
		d.text(353, d.y, 100, money(r.unit), 10)
		d.text(463, d.y, 92, money(r.unit*r.count), 10)
		d.y += 20
	}
	d.pair("Turistas", fmt.Sprintf("%d en total / %d cupos computables", p.NationalAdults+p.ForeignAdults+len(p.Minors), p.CapacityCount))
	d.pair("Importe de la compra", money(p.TotalCents))
	method := "QR"
	if p.PaymentMethod == "transfer" {
		method = "Transferencia bancaria"
	}
	d.pair("Medio de pago", method)
	d.section("ESTADO AL GENERAR ESTE DOCUMENTO")
	status := map[string]string{models.PurchaseConfirmed: "Pago confirmado", models.PurchaseRefundPending: "Reembolso pendiente", models.PurchaseRefunded: "Reembolso realizado"}[p.Status]
	d.pair("Compra", status)
	if p.PaymentConfirmedAt != nil {
		d.pair("Pago confirmado el", date(*p.PaymentConfirmedAt))
	} else {
		d.paragraph("Fecha de confirmación del pago: sin registro histórico.")
	}
	operational := "Pendiente de autorización de la agencia"
	switch {
	case departure.Status == "cancelled":
		operational = "Salida cancelada"
	case departure.Status == "completed" || !departure.StartsAt.After(now):
		operational = "Salida pasada"
	case !agencyActive:
		operational = "Agencia no habilitada para operar la salida"
	case slices.Contains([]string{"open", "confirmed", "closed"}, departure.Status) && departure.AuthorizedAt != nil && departure.AuthorizationRevokedAt == nil && departure.ConfirmedCapacity >= departure.MinCapacity:
		operational = "Salida autorizada por la agencia"
	}
	d.pair("Situación de la salida", operational)
	if p.Status == models.PurchaseRefundPending || p.Status == models.PurchaseRefunded {
		d.pair("Motivo de devolución", p.RefundReason)
		if p.Status == models.PurchaseRefundPending && p.RefundDueAt != nil {
			d.pair("Plazo de devolución", date(*p.RefundDueAt))
		}
		if p.RefundedAt != nil {
			d.pair("Devolución registrada", date(*p.RefundedAt))
		}
		if p.RefundReference != "" {
			d.pair("Referencia de devolución", p.RefundReference)
		}
	}
	d.section("CONDICIONES E INDICACIONES")
	if s.CancellationAllowed {
		d.paragraph(fmt.Sprintf("Cancelación por el turista: aviso mínimo de %d horas antes de la salida, con devolución del 100 %% según la política de la compra.", s.CancellationNoticeHours))
	} else {
		d.paragraph("Esta compra no admite cancelación voluntaria por el turista según la política contratada.")
	}
	if len(s.Bring) > 0 {
		d.pair("Llevar", strings.Join(s.Bring, ", "))
	}
	d.paragraph("El pago confirmado y la autorización de la salida son decisiones independientes. Consulta el estado vigente en Mis compras y coordina con la agencia.")
	d.paragraph("Constancia interna de compra. No es una factura fiscal. El respaldo bancario se conserva por separado en la operación.")
	if s.Origin == "migracion" {
		d.paragraph("Compra anterior a esta función: datos comerciales recuperados el " + date(s.CapturedAt) + "; no se dispone de una copia original anterior del catálogo.")
	}
	if d.err != nil {
		return nil, d.err
	}
	var output bytes.Buffer
	if _, err = d.pdf.WriteTo(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
func nonempty(values ...string) []string {
	out := []string{}
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}
