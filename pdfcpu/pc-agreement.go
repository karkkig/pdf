package pcpdf

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// AgreementData holds all fields for the Services Agreement.
type AgreementData struct {
	State           string
	Day             string
	Month           string
	Year            string
	ProviderName    string
	ProviderAddress string
	BuyerName       string
	BuyerAddress    string
	Services        []ServiceItem
	TotalPrice      string
	TaxResponsible  string // "Service Provider" or "Buyer"
	PaymentMethod   string
	// Page 2 fields
	PayerName    string
	PayeeName    string
	PayerUPI     string
	Interval     string
	TotalAmount  string
	PaymentDates []string
	LateFee      string
	BounceFee    string
	LenderAction string
	Terms        string
}

type ServiceItem struct {
	Description string
	NumProjects string
	PricePerPrj string
}

// GenerateAgreement creates a two-page services agreement PDF.
func GenerateAgreement(outPath string) error {
	data := AgreementData{
		State:           "California",
		Day:             "1st",
		Month:           "January",
		Year:            "2030",
		ProviderName:    "Studio Shodwe LLC",
		ProviderAddress: "123 Anywhere St., Any City, ST 12345",
		BuyerName:       "Rachel Beaudry",
		BuyerAddress:    "456 Other St., Other City, ST 67890",
		Services: []ServiceItem{
			{"Brand Identity Design", "2", "$500"},
			{"Website Mockup", "1", "$800"},
			{"Social Media Kit", "3", "$250"},
		},
		TotalPrice:     "2,300.00",
		TaxResponsible: "Service Provider",
		PaymentMethod:  "Wire transfer",
		PayerName:      "Rachel Beaudry",
		PayeeName:      "Studio Shodwe LLC",
		PayerUPI:       "studioghodwe@bank",
		Interval:       "monthly",
		TotalAmount:    "$2,300.00",
		PaymentDates: []string{
			"Feb 1, 2030 — $384",
			"Mar 1, 2030 — $384",
			"Apr 1, 2030 — $383",
			"May 1, 2030 — $383",
			"Jun 1, 2030 — $383",
			"Jul 1, 2030 — $383",
		},
		LateFee:      "$25",
		BounceFee:    "$50",
		LenderAction: "contacting a debt collection service",
		Terms:        "Governing law: State of California. Dispute resolution: Arbitration.",
	}

	page1 := buildAgreementPage1(data)
	page2 := buildAgreementPage2(data)

	raw := buildRawPDF([]string{page1, page2}, 595, 842)
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

func buildAgreementPage1(d AgreementData) string {
	var buf bytes.Buffer
	w, h := 595.0, 842.0

	text := func(x, y float64, size int, bold bool, color, s string) {
		fontName := "Helvetica"
		if bold {
			fontName = "Helvetica-Bold"
		}
		fmt.Fprintf(&buf, "BT /%s %d Tf %s rg %.2f %.2f Td (%s) Tj ET\n",
			fontName, size, color, x, y, pdfEscape(s))
	}
	line := func(x1, y1, x2, y2, width float64, color string) {
		fmt.Fprintf(&buf, "%s RG %.2f w %.2f %.2f m %.2f %.2f l S\n",
			color, width, x1, y1, x2, y2)
	}
	rect := func(x, y, rw, rh float64, fill string) {
		fmt.Fprintf(&buf, "%s rg %.2f %.2f %.2f %.2f re f\n", fill, x, y, rw, rh)
	}
	checkbox := func(x, y float64, checked bool) {
		fmt.Fprintf(&buf, "0 0 0 RG 0.5 w %.2f %.2f 8 8 re S\n", x, y)
		if checked {
			fmt.Fprintf(&buf, "0 0 0 rg BT /Helvetica 8 Tf %.2f %.2f Td (x) Tj ET\n", x+2, y+1)
		}
	}

	// ── State line ──────────────────────────────────────────────────────────
	text(50, h-50, 9, false, "0 0 0", fmt.Sprintf("State of %s", d.State))

	// ── Title ───────────────────────────────────────────────────────────────
	text(160, h-75, 20, true, "0 0 0", "SERVICES AGREEMENT")
	line(50, h-82, w-50, h-82, 2, "0 0 0")
	line(50, h-85, w-50, h-85, 0.5, "0 0 0")

	// ── Preamble ─────────────────────────────────────────────────────────────
	preamble := fmt.Sprintf(
		"This Services Agreement (this \"Agreement\") is entered into as of the %s day of %s, 20%s, by and among/between:",
		d.Day, d.Month, d.Year[2:])
	wrapText(&buf, preamble, 50, h-105, w-100, 9, false, "0 0 0")

	// ── Service Provider ──────────────────────────────────────────────────────
	text(50, h-140, 9, true, "0 0 0", "Service Provider(s):")
	text(180, h-140, 9, false, "0 0 0", d.ProviderName+" [Name], located at")
	text(50, h-154, 9, false, "0 0 0", d.ProviderAddress+` (collectively "Service Provider") and`)

	// ── Buyer ────────────────────────────────────────────────────────────────
	text(50, h-175, 9, true, "0 0 0", "Buyer(s):")
	text(120, h-175, 9, false, "0 0 0", d.BuyerName+" [Name], located at")
	text(50, h-189, 9, false, "0 0 0", d.BuyerAddress+` (collectively "Buyer").`)

	text(50, h-210, 9, false, "0 0 0",
		`Each Service Provider and Buyer may be referred to in this Agreement individually as a "Party" and collectively as the "Parties."`)

	// ── Section 1: Services ──────────────────────────────────────────────────
	text(50, h-235, 9, true, "0 0 0", "1. Services.")
	text(105, h-235, 9, false, "0 0 0",
		"Service Provider agrees to provide and Buyer agrees to purchase the following services for the specific projects described below:")

	// Table header
	tableTop := h - 265.0
	rect(50, tableTop, w-100, 20, "0.15 0.15 0.15")
	text(55, tableTop+6, 9, true, "1 1 1", "Description of Services")
	text(355, tableTop+6, 9, true, "1 1 1", "Number of Projects")
	text(460, tableTop+6, 9, true, "1 1 1", "Price per Project")

	rowH := 18.0
	for i, svc := range d.Services {
		ry := tableTop - float64(i+1)*rowH
		if i%2 == 0 {
			rect(50, ry, w-100, rowH, "0.93 0.93 0.93")
		}
		text(55, ry+5, 8, false, "0 0 0", svc.Description)
		text(385, ry+5, 8, false, "0 0 0", svc.NumProjects)
		text(475, ry+5, 8, false, "0 0 0", svc.PricePerPrj)
	}
	// Extra blank rows
	for i := len(d.Services); i < 6; i++ {
		ry := tableTop - float64(i+1)*rowH
		if i%2 == 0 {
			rect(50, ry, w-100, rowH, "0.97 0.97 0.97")
		}
		fmt.Fprintf(&buf, "0.7 0.7 0.7 RG 0.3 w %.2f %.2f m %.2f %.2f l S\n",
			50.0, ry, w-50, ry)
	}

	tableBot := tableTop - 6*rowH
	fmt.Fprintf(&buf, "0 0 0 RG 0.5 w %.2f %.2f %.2f %.2f re S\n",
		50.0, tableBot, w-100, tableTop-tableBot)

	// ── Section 2: Purchase Price ─────────────────────────────────────────────
	s2y := tableBot - 30
	text(50, s2y, 9, true, "0 0 0", "2. Purchase Price.")
	text(145, s2y, 9, false, "0 0 0",
		fmt.Sprintf("Buyer will pay to Service Provider and for all obligations specified in this Agreement, if any, as the full and complete purchase price, the sum of $%s.", d.TotalPrice))

	taxY := s2y - 20
	text(50, taxY, 9, false, "0 0 0", "Unless otherwise stated, (Check one)")
	checkbox(240, taxY-2, d.TaxResponsible == "Service Provider")
	text(252, taxY, 9, false, "0 0 0", "Service Provider")
	checkbox(340, taxY-2, d.TaxResponsible == "Buyer")
	text(352, taxY, 9, false, "0 0 0", "Buyer shall be responsible for all taxes in connection with the purchase of Services in this Agreement.")

	// ── Section 3: Payment ────────────────────────────────────────────────────
	s3y := taxY - 30
	text(50, s3y, 9, true, "0 0 0", "3. Payment.")
	text(105, s3y, 9, false, "0 0 0", "Payment for the Services will be by: (Check one)")

	payOpts := []string{"Cash", "Personal check", "Cashier's check", "Money order",
		"Credit or debit card", "Wire transfer", "Other: _______________"}
	for i, opt := range payOpts {
		col := 0.0
		row := i
		if i >= 4 {
			col = 250
			row = i - 4
		}
		cy2 := s3y - 18 - float64(row)*16
		checkbox(55+col, cy2-2, opt == d.PaymentMethod)
		text(67+col, cy2, 9, false, "0 0 0", opt)
	}

	// ── Footer ────────────────────────────────────────────────────────────────
	line(50, 30, w-50, 30, 0.3, "0.7 0.7 0.7")
	text(50, 18, 8, false, "0.5 0.5 0.5", "Page 1 of 2 — Services Agreement")

	return buf.String()
}

func buildAgreementPage2(d AgreementData) string {
	var buf bytes.Buffer
	w, h := 595.0, 842.0

	text := func(x, y float64, size int, bold bool, color, s string) {
		fontName := "Helvetica"
		if bold {
			fontName = "Helvetica-Bold"
		}
		fmt.Fprintf(&buf, "BT /%s %d Tf %s rg %.2f %.2f Td (%s) Tj ET\n",
			fontName, size, color, x, y, pdfEscape(s))
	}
	line := func(x1, y1, x2, y2, width float64, color string) {
		fmt.Fprintf(&buf, "%s RG %.2f w %.2f %.2f m %.2f %.2f l S\n",
			color, width, x1, y1, x2, y2)
	}

	// ── Header ────────────────────────────────────────────────────────────────
	text(50, h-50, 9, false, "0.4 0.4 0.4",
		fmt.Sprintf("Services Agreement (cont.) — %s & %s", d.ProviderName, d.BuyerName))
	line(50, h-58, w-50, h-58, 0.5, "0.7 0.7 0.7")

	// ── Payment Plan section ──────────────────────────────────────────────────
	cy := h - 90.0

	p2text := fmt.Sprintf(
		"By this contract, {%s} agrees to make payments to {%s}, hereafter known as \"Lender,\" by the "+
			"following schedule in exchange for {services rendered}. This payment schedule is enforceable "+
			"by law, and the methods described below will be use in cases of delinquent payment.",
		d.PayerName, d.PayeeName)
	cy = wrapTextRet(&buf, p2text, 50, cy, w-100, 9, false, "0 0 0")

	cy -= 14
	text(50, cy, 9, false, "0 0 0",
		fmt.Sprintf("By this agreement, it is agreed that a payment of {amount} will be surrendered to the Lender every {%s} until", d.Interval))
	cy -= 13
	text(50, cy, 9, false, "0 0 0",
		fmt.Sprintf("the total of the payment required, which is %s, has been delivered. The payment plan will take the following form:", d.TotalAmount))
	cy -= 18

	// Bullet list of payment dates
	for _, pd := range d.PaymentDates {
		fmt.Fprintf(&buf, "0 0 0 rg %.2f %.2f 3 3 re f\n", 57.0, cy+2)
		text(65, cy, 9, false, "0 0 0", pd)
		cy -= 14
	}

	cy -= 10
	text(50, cy, 9, false, "0 0 0", "These payments include any interest and other charges that may apply.")
	cy -= 20

	body2 := fmt.Sprintf(
		"This agreement is binding, and failure to meet its terms will allow the Lender to take certain recourse. "+
			"First, late payments will incur a fee of %s every {%s} interval. Insufficient payment and bounced checks will "+
			"incur a fee of %s. If payment should not be delivered at all, Lender will be entitled to {%s}.",
		d.LateFee, d.Interval, d.BounceFee, d.LenderAction)
	cy = wrapTextRet(&buf, body2, 50, cy, w-100, 9, false, "0 0 0")

	cy -= 14
	text(50, cy, 9, false, "0 0 0",
		fmt.Sprintf("In addition, the following terms and conditions apply: {%s}.", d.Terms))
	cy -= 20

	body3 := "By signing this agreement, all parties agree to the terms as described above. Alterations to this agreement can only be made by both parties and must be placed in writing. Both parties will receive a printed copy of this agreement, and will be responsible for upholding its terms."
	cy = wrapTextRet(&buf, body3, 50, cy, w-100, 9, false, "0 0 0")

	// ── Signature lines ───────────────────────────────────────────────────────
	sigY := cy - 50.0
	line(50, sigY, 280, sigY, 0.5, "0 0 0")
	line(320, sigY, 545, sigY, 0.5, "0 0 0")
	text(50, sigY-14, 9, false, "0 0 0", fmt.Sprintf("(%s — Payer)", d.PayerName))
	text(320, sigY-14, 9, false, "0 0 0", "(Date)")

	sig2Y := sigY - 50.0
	line(50, sig2Y, 280, sig2Y, 0.5, "0 0 0")
	line(320, sig2Y, 545, sig2Y, 0.5, "0 0 0")
	text(50, sig2Y-14, 9, false, "0 0 0", fmt.Sprintf("(%s — Payee)", d.PayeeName))
	text(320, sig2Y-14, 9, false, "0 0 0", "(Date)")

	// ── Footer ────────────────────────────────────────────────────────────────
	line(50, 30, w-50, 30, 0.3, "0.7 0.7 0.7")
	text(50, 18, 8, false, "0.5 0.5 0.5", "Page 2 of 2 — Services Agreement")

	return buf.String()
}

// wrapText writes multi-line text by breaking at word boundaries.
func wrapText(buf *bytes.Buffer, s string, x, y, maxW float64, size int, bold bool, color string) {
	wrapTextRet(buf, s, x, y, maxW, size, bold, color)
}

// wrapTextRet is like wrapText but returns the final y position.
func wrapTextRet(buf *bytes.Buffer, s string, x, y, maxW float64, size int, bold bool, color string) float64 {
	fontName := "Helvetica"
	if bold {
		fontName = "Helvetica-Bold"
	}
	// Rough char width for proportional font at given size
	charW := float64(size) * 0.52
	charsPerLine := int(maxW / charW)

	words := splitWords(s)
	line := ""
	curY := y
	for _, w := range words {
		candidate := line
		if candidate != "" {
			candidate += " "
		}
		candidate += w
		if len(candidate) > charsPerLine && line != "" {
			fmt.Fprintf(buf, "BT /%s %d Tf %s rg %.2f %.2f Td (%s) Tj ET\n",
				fontName, size, color, x, curY, pdfEscape(line))
			curY -= float64(size) + 3
			line = w
		} else {
			line = candidate
		}
	}
	if line != "" {
		fmt.Fprintf(buf, "BT /%s %d Tf %s rg %.2f %.2f Td (%s) Tj ET\n",
			fontName, size, color, x, curY, pdfEscape(line))
		curY -= float64(size) + 3
	}
	return curY
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\n' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}
