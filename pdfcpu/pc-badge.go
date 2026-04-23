package pcpdf

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// BadgeData holds fields for the GPay-style payment badge.
type BadgeData struct {
	PaymentMethod string // e.g. "GPay"
	BusinessName  string
	PhoneNumber   string
	UPIAddress    string // shown below QR
}

// GeneratePaymentBadge creates a vertical payment badge PDF (portrait card).
func GeneratePaymentBadge(outPath string) error {
	data := BadgeData{
		PaymentMethod: "GPay",
		BusinessName:  "Your Business Name",
		PhoneNumber:   "+91 12345 67890",
		UPIAddress:    "12345 67890@yhh",
	}

	w, h := 250.0, 400.0 // narrow card dimensions in points
	var buf bytes.Buffer

	// ── helpers ────────────────────────────────────────────────────────────────
	text := func(x, y float64, size int, bold bool, color, s string) {
		fontName := "Helvetica"
		if bold {
			fontName = "Helvetica-Bold"
		}
		fmt.Fprintf(&buf, "BT /%s %d Tf %s rg %.2f %.2f Td (%s) Tj ET\n",
			fontName, size, color, x, y, pdfEscape(s))
	}
	rect := func(x, y, rw, rh float64, fill string) {
		fmt.Fprintf(&buf, "%s rg %.2f %.2f %.2f %.2f re f\n", fill, x, y, rw, rh)
	}
	line := func(x1, y1, x2, y2 float64, width float64, color string) {
		fmt.Fprintf(&buf, "%s RG %.2f w %.2f %.2f m %.2f %.2f l S\n",
			color, width, x1, y1, x2, y2)
	}

	// ── White background ──────────────────────────────────────────────────────
	rect(0, 0, w, h, "1 1 1")

	// ── Light green header band ───────────────────────────────────────────────
	rect(0, h-90, w, 90, "0.85 0.95 0.85")

	// ── GPay logo placeholder ─────────────────────────────────────────────────
	// Simulate the "G" with a coloured circle
	cx, cy, r := w/2, h-50.0, 22.0
	drawCircleApprox(&buf, cx, cy, r, "0.26 0.52 0.96") // blue
	text(cx-7, cy-5, 14, true, "1 1 1", "G")

	// ── Payment label ─────────────────────────────────────────────────────────
	text(w/2-22, h-82, 8, false, "0.2 0.2 0.2", data.PaymentMethod)
	text(w/2-30, h-94, 8, false, "0.4 0.4 0.4", "accepted here")

	// ── Divider ───────────────────────────────────────────────────────────────
	line(20, h-105, w-20, h-105, 0.5, "0.8 0.8 0.8")

	// ── Business details ──────────────────────────────────────────────────────
	text(w/2-50, h-125, 11, true, "0 0 0", data.BusinessName)
	text(w/2-42, h-142, 9, false, "0.3 0.3 0.3", data.PhoneNumber)

	// ── QR code placeholder (drawn as a bordered square with pattern) ─────────
	qx, qy, qs := 45.0, 130.0, 160.0
	rect(qx, qy, qs, qs, "1 1 1")
	// Outer border
	fmt.Fprintf(&buf, "0 0 0 RG 2 w %.2f %.2f %.2f %.2f re S\n", qx, qy, qs, qs)

	// Position markers (three corner squares — typical QR style)
	drawQRCorner(&buf, qx+8, qy+qs-36, 28)
	drawQRCorner(&buf, qx+qs-36, qy+qs-36, 28)
	drawQRCorner(&buf, qx+8, qy+8, 28)

	// Data modules simulation (small squares scattered)
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if (row+col)%3 != 0 {
				mx := qx + 45 + float64(col)*9
				my := qy + 45 + float64(row)*8
				rect(mx, my, 6, 6, "0 0 0")
			}
		}
	}

	// ── UPI address below QR ──────────────────────────────────────────────────
	text(w/2-50, qy-20, 9, false, "0.3 0.3 0.3", data.UPIAddress)

	// ── Bottom accent bar ─────────────────────────────────────────────────────
	rect(0, 0, w, 15, "0.26 0.52 0.96")

	raw := buildRawPDF([]string{buf.String()}, w, h)
	rs := bytes.NewReader([]byte(raw))
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return api.Optimize(rs, outFile, conf)
}

// drawCircleApprox approximates a circle using 4 Bézier curves.
func drawCircleApprox(buf *bytes.Buffer, cx, cy, r float64, fill string) {
	k := 0.5523 * r
	fmt.Fprintf(buf, "%s rg\n", fill)
	fmt.Fprintf(buf, "%.2f %.2f m\n", cx+r, cy)
	fmt.Fprintf(buf, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		cx+r, cy+k, cx+k, cy+r, cx, cy+r)
	fmt.Fprintf(buf, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		cx-k, cy+r, cx-r, cy+k, cx-r, cy)
	fmt.Fprintf(buf, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		cx-r, cy-k, cx-k, cy-r, cx, cy-r)
	fmt.Fprintf(buf, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		cx+k, cy-r, cx+r, cy-k, cx+r, cy)
	buf.WriteString("f\n")
}

// drawQRCorner draws the three-square finder pattern used in QR codes.
func drawQRCorner(buf *bytes.Buffer, x, y, size float64) {
	inner := size * 0.6
	innerOff := (size - inner) / 2
	core := size * 0.3
	coreOff := (size - core) / 2

	fmt.Fprintf(buf, "0 0 0 rg %.2f %.2f %.2f %.2f re f\n", x, y, size, size)
	fmt.Fprintf(buf, "1 1 1 rg %.2f %.2f %.2f %.2f re f\n",
		x+innerOff, y+innerOff, inner, inner)
	fmt.Fprintf(buf, "0 0 0 rg %.2f %.2f %.2f %.2f re f\n",
		x+coreOff, y+coreOff, core, core)
}
