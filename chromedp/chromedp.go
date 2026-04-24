package cdpdf

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ─── Data structs ────────────────────────────────────────────────────────────

type PartyInfo struct {
	Name    string
	Address string
	City    string
	Phone   string
	Email   string
}

type InvoiceItem struct {
	Description string
	UnitPrice   float64
	Qty         int
}

func (i InvoiceItem) Amount() float64 { return i.UnitPrice * float64(i.Qty) }

type PaymentInfo struct {
	Bank          string
	AccountName   string
	AccountNumber string
}

type PreparedByInfo struct {
	Name  string
	Title string
}

type InvoiceData struct {
	Number        string
	Date          time.Time
	Provider      PartyInfo
	Client        PartyInfo
	Items         []InvoiceItem
	TaxRate       float64
	Notes         string
	PaymentMethod PaymentInfo
	PreparedBy    PreparedByInfo
}

func (d InvoiceData) SubTotal() float64 {
	var t float64
	for _, it := range d.Items {
		t += it.Amount()
	}
	return t
}
func (d InvoiceData) TaxAmount() float64   { return d.SubTotal() * d.TaxRate }
func (d InvoiceData) TotalAmount() float64 { return d.SubTotal() + d.TaxAmount() }

// ── Agreement (page 1) ───────────────────────────────────────────────────────

type ServiceItem struct {
	Description string
	NumProjects string
	PricePerPrj string
}

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
	PurchasePrice   string
	Notes           string
}

// ── Payment plan (page 2) ────────────────────────────────────────────────────

type PaymentEntry struct {
	Date   string
	Amount string
}

type PaymentPlanData struct {
	Payer           string
	Payee           string
	Product         string
	AmountPerPeriod string
	Interval        string
	TotalAmount     string
	Payments        []PaymentEntry
	LateFee         string
	BounceFee       string
	LenderAction    string
	TermsConditions string
}

// FullAgreementData combines both pages into one template execution.
type FullAgreementData struct {
	AgreementData
	PaymentPlanData
}

// ── GPay badge ───────────────────────────────────────────────────────────────

type GPayBadgeData struct {
	BusinessName string
	PhoneNumber  string
	UPIHandle    string
	QRContent    string // used to build QR URL
	QRCodePath   string // optional local image path (overrides QRContent)
}

func (d GPayBadgeData) QRURL() string {
	if d.QRCodePath != "" {
		return d.QRCodePath
	}
	return "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" +
		url.QueryEscape(d.QRContent)
}

// ─── Template rendering ───────────────────────────────────────────────────────

func ReadHTML(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func RenderHTMLTemplate(path string, data any) (string, error) {
	funcs := template.FuncMap{
		"money": func(v float64) string {
			return fmt.Sprintf("$%.2f", v)
		},
		"percent": func(v float64) string {
			return fmt.Sprintf("%.0f%%", v*100)
		},
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"nl2br": func(s string) template.HTML {
			escaped := template.HTMLEscapeString(s)
			return template.HTML(strings.ReplaceAll(escaped, "\n", "<br>"))
		},
	}

	tmpl, err := template.New(filepath.Base(path)).Funcs(funcs).ParseFiles(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ─── PDF generation ───────────────────────────────────────────────────────────

func GeneratePDF(html string, output string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var pdfBuf []byte

	formatted := `<html><head><meta charset="UTF-8"><style>body{margin:0;}</style></head><body>` +
		html + `</body></html>`

	htmlURL := "data:text/html," + url.PathEscape(formatted)

	err := chromedp.Run(ctx,
		chromedp.Navigate(htmlURL),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				Do(ctx)
			pdfBuf = buf
			return err
		}),
	)
	if err != nil {
		return err
	}
	return os.WriteFile(output, pdfBuf, 0644)
}
