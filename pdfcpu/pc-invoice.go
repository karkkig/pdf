package pcpdf

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// InvoiceData holds all invoice fields
type InvoiceData struct {
	InvoiceNumber string
	InvoiceDate   string
	CompanyName   string
	CompanyAddr   string
	CompanyPhone  string
	CompanyEmail  string
	BillToName    string
	BillToAddr    string
	BillToPhone   string
	BillToEmail   string
	Items         []InvoiceItem
	SubTotal      float64
	TaxRate       float64
	TaxAmount     float64
	Total         float64
	Notes         string
	BankName      string
	AccountName   string
	AccountNumber string
	PreparedBy    string
	PreparedTitle string
}

type InvoiceItem struct {
	Description string
	UnitPrice   float64
	Qty         int
	Amount      float64
}

// GenerateInvoice creates an invoice PDF similar to the Studio Shodwe template.
func GenerateInvoice(outPath string) error {
	data := InvoiceData{
		InvoiceNumber: "#123456",
		InvoiceDate:   "24/05/2030",
		CompanyName:   "STUDIO SHODWE",
		CompanyAddr:   "123 Anywhere St., Any City, ST 12345",
		CompanyPhone:  "+123-456-7890",
		CompanyEmail:  "hello@reallygreatsite.com",
		BillToName:    "Rachel Beaudry",
		BillToAddr:    "123 Anywhere St., Any City, ST 12345",
		BillToPhone:   "+123-456-7890",
		BillToEmail:   "hello@reallygreatsite.com",
		Items: []InvoiceItem{
			{"Service 1", 100.00, 1, 100.00},
			{"Service 2", 150.00, 1, 150.00},
			{"Service 3", 200.00, 1, 200.00},
		},
		SubTotal:      450.00,
		TaxRate:       6,
		TaxAmount:     36.00,
		Total:         486.00,
		Notes:         "Payment is due within 15 days of receiving this invoice.",
		BankName:      "Borcelle Bank",
		AccountName:   "Studio Shodwe",
		AccountNumber: "1234567890",
		PreparedBy:    "Benjamin Shah",
		PreparedTitle: "Sales Administrator, Studio Shodwe",
	}

	var buf bytes.Buffer
	w, h := 595.0, 842.0 // A4 in points

	// ── Helpers ────────────────────────────────────────────────────────────────
	text := func(x, y float64, size int, bold bool, color, s string) {
		fontName := "Helvetica"
		if bold {
			fontName = "Helvetica-Bold"
		}
		fmt.Fprintf(&buf, "BT /%s %d Tf %s rg %.2f %.2f Td (%s) Tj ET\n",
			fontName, size, color, x, y, pdfEscape(s))
	}
	line := func(x1, y1, x2, y2 float64, width float64, color string) {
		fmt.Fprintf(&buf, "%s RG %.2f w %.2f %.2f m %.2f %.2f l S\n",
			color, width, x1, y1, x2, y2)
	}
	rect := func(x, y, rw, rh float64, fill string) {
		fmt.Fprintf(&buf, "%s rg %.2f %.2f %.2f %.2f re f\n",
			fill, x, y, rw, rh)
	}

	// ── Purple accent bar top-right ─────────────────────────────────────────
	rect(w-130, h-80, 60, 60, "0.38 0.20 0.60")
	text(w-122, h-40, 8, true, "1 1 1", "STUDIO")
	text(w-122, h-52, 8, true, "1 1 1", "SHODWE")

	// ── Title ───────────────────────────────────────────────────────────────
	text(50, h-60, 28, true, "0 0 0", "INVOICE")
	text(50, h-78, 9, false, "0.4 0.4 0.4", fmt.Sprintf("Invoice Number: %s", data.InvoiceNumber))
	text(50, h-90, 9, false, "0.4 0.4 0.4", fmt.Sprintf("Invoice Date: %s", data.InvoiceDate))

	// ── Divider ─────────────────────────────────────────────────────────────
	line(50, h-105, w-50, h-105, 0.5, "0.7 0.7 0.7")

	// ── From / Bill To ──────────────────────────────────────────────────────
	text(50, h-125, 9, true, "0.38 0.20 0.60", data.CompanyName)
	text(50, h-138, 8, false, "0.3 0.3 0.3", data.CompanyAddr)
	text(50, h-150, 8, false, "0.3 0.3 0.3", data.CompanyPhone)
	text(50, h-162, 8, false, "0.3 0.3 0.3", data.CompanyEmail)

	text(300, h-118, 9, true, "0.38 0.20 0.60", "BILL TO")
	text(300, h-130, 9, false, "0 0 0", data.BillToName)
	text(300, h-142, 8, false, "0.3 0.3 0.3", data.BillToAddr)
	text(300, h-154, 8, false, "0.3 0.3 0.3", data.BillToPhone)
	text(300, h-166, 8, false, "0.3 0.3 0.3", data.BillToEmail)

	// ── Table header ─────────────────────────────────────────────────────────
	tableTop := h - 200.0
	rect(50, tableTop-15, w-100, 20, "0.15 0.15 0.15")
	// FIX: "& " is plain ASCII — safe. No unicode issues here.
	text(55, tableTop-10, 9, true, "1 1 1", "Item & Description")
	text(360, tableTop-10, 9, true, "1 1 1", "Unit Price")
	text(440, tableTop-10, 9, true, "1 1 1", "Qty")
	text(490, tableTop-10, 9, true, "1 1 1", "Amount")

	// ── Table rows ────────────────────────────────────────────────────────────
	rowH := 22.0
	for i, item := range data.Items {
		y := tableTop - 15 - float64(i+1)*rowH
		if i%2 == 0 {
			rect(50, y, w-100, rowH, "0.95 0.95 0.95")
		}
		text(55, y+6, 9, false, "0 0 0", item.Description)
		text(360, y+6, 9, false, "0 0 0", fmt.Sprintf("$%.2f", item.UnitPrice))
		text(444, y+6, 9, false, "0 0 0", fmt.Sprintf("%d", item.Qty))
		text(490, y+6, 9, false, "0 0 0", fmt.Sprintf("$%.2f", item.Amount))
	}
	line(50, tableTop-15-float64(len(data.Items))*rowH, w-50,
		tableTop-15-float64(len(data.Items))*rowH, 0.5, "0.7 0.7 0.7")

	// ── Notes / Terms ─────────────────────────────────────────────────────────
	notesY := tableTop - 15 - float64(len(data.Items))*rowH - 30
	text(50, notesY, 9, true, "0 0 0", "NOTES / TERMS:")
	// Use data.Notes so the field drives the output; \n in the string is split
	// into two separate text calls to avoid the newline going into the PDF literal.
	noteLines := splitLines(data.Notes)
	for i, nl := range noteLines {
		text(50, notesY-14-float64(i)*12, 8, false, "0.4 0.4 0.4", nl)
	}

	// ── Totals box ────────────────────────────────────────────────────────────
	totX := 360.0
	totY := notesY + 10
	text(totX, totY, 9, false, "0.3 0.3 0.3", "Sub-Total")
	text(490, totY, 9, false, "0 0 0", fmt.Sprintf("$%.2f", data.SubTotal))
	// FIX: use %% to produce a literal % inside Sprintf; no unicode needed.
	text(totX, totY-16, 9, false, "0.3 0.3 0.3", fmt.Sprintf("Tax (%.0f%%)", data.TaxRate))
	text(490, totY-16, 9, false, "0 0 0", fmt.Sprintf("$%.2f", data.TaxAmount))
	line(totX, totY-24, w-50, totY-24, 1, "0.38 0.20 0.60")
	rect(totX, totY-44, w-50-totX, 20, "0.38 0.20 0.60")
	text(totX+5, totY-38, 10, true, "1 1 1", "Total")
	text(totX+100, totY-38, 10, true, "1 1 1", fmt.Sprintf("$%.2f", data.Total))

	// ── Payment Method / Prepared By ─────────────────────────────────────────
	pmY := 100.0
	line(50, pmY+55, w-50, pmY+55, 0.5, "0.7 0.7 0.7")
	text(50, pmY+40, 9, true, "0 0 0", "PAYMENT METHOD")
	text(50, pmY+26, 8, false, "0.4 0.4 0.4", fmt.Sprintf("Bank: %s", data.BankName))
	text(50, pmY+14, 8, false, "0.4 0.4 0.4", fmt.Sprintf("Account Name: %s", data.AccountName))
	text(50, pmY+2, 8, false, "0.4 0.4 0.4", fmt.Sprintf("Account Number: %s", data.AccountNumber))

	text(300, pmY+40, 9, true, "0 0 0", "PREPARED BY")
	text(300, pmY+26, 9, false, "0 0 0", data.PreparedBy)
	text(300, pmY+14, 8, false, "0.4 0.4 0.4", data.PreparedTitle)

	return writeSinglePagePDF(outPath, buf.String(), w, h)
}

// writeSinglePagePDF writes a raw content stream as a single-page PDF via pdfcpu.
func writeSinglePagePDF(outPath, contentStream string, w, h float64) error {
	raw := buildRawPDF([]string{contentStream}, w, h)
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

// buildRawPDF constructs a minimal valid PDF byte sequence for one or more pages.
func buildRawPDF(pages []string, w, h float64) string {
	var b bytes.Buffer

	b.WriteString("%PDF-1.4\n")

	obj1Off := b.Len()
	b.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	obj2Off := b.Len()
	pageCount := len(pages)
	kidsRef := ""
	for i := range pages {
		kidsRef += fmt.Sprintf("%d 0 R ", 3+i*2)
	}
	fmt.Fprintf(&b, "2 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n",
		kidsRef, pageCount)

	pageOffsets := make([]int, pageCount)
	contentOffsets := make([]int, pageCount)

	for i, cs := range pages {
		contentObj := 3 + i*2 + 1
		pageOffsets[i] = b.Len()
		fmt.Fprintf(&b,
			"%d 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f]\n"+
				"   /Resources << /Font << /Helvetica << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n"+
				"                             /Helvetica-Bold << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >> >> >>\n"+
				"   /Contents %d 0 R >>\nendobj\n",
			3+i*2, w, h, contentObj)

		contentOffsets[i] = b.Len()
		csBytes := []byte(cs)
		fmt.Fprintf(&b,
			"%d 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n",
			contentObj, len(csBytes), cs)
	}

	xrefOffset := b.Len()
	totalObjs := 2 + pageCount*2 + 1

	fmt.Fprintf(&b, "xref\n0 %d\n", totalObjs)
	b.WriteString("0000000000 65535 f \n")
	fmt.Fprintf(&b, "%010d 00000 n \n", obj1Off)
	fmt.Fprintf(&b, "%010d 00000 n \n", obj2Off)
	for i := range pages {
		fmt.Fprintf(&b, "%010d 00000 n \n", pageOffsets[i])
		fmt.Fprintf(&b, "%010d 00000 n \n", contentOffsets[i])
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		totalObjs, xrefOffset)

	return b.String()
}

// pdfEscape escapes a UTF-8 string for use inside a PDF literal string with
// WinAnsiEncoding (Windows-1252), which is the encoding of the built-in
// Type1 fonts (Helvetica, Times-Roman, Courier, etc.).
//
// Rules:
//   - ( ) \ are backslash-escaped (PDF spec requirement).
//   - Printable ASCII 0x20–0x7E passes through unchanged.
//   - Unicode code points that exist in WinAnsi outside ASCII are written as
//     PDF octal escapes \NNN so the correct byte reaches the font.
//   - Truly unmappable runes fall back to a plain hyphen (for dash variants)
//     or '?'.
