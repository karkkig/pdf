package mapdf

import (
	"fmt"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// ─── Data types ────────────────────────────────────────────────────────────────

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

func (d InvoiceData) SubTotal() float64 {
	var s float64
	for _, it := range d.Items {
		s += it.Amount()
	}
	return s
}
func (d InvoiceData) Tax() float64   { return d.SubTotal() * d.TaxRate }
func (d InvoiceData) Total() float64 { return d.SubTotal() + d.Tax() }

// ─── Colors ────────────────────────────────────────────────────────────────────

var (
	invBlack     = props.Color{Red: 20, Green: 20, Blue: 20}
	invWhite     = props.Color{Red: 255, Green: 255, Blue: 255}
	invDarkGray  = props.Color{Red: 80, Green: 80, Blue: 80}
	invLightGray = props.Color{Red: 245, Green: 245, Blue: 245}
	invBorder    = props.Color{Red: 180, Green: 180, Blue: 180}
	invTotalBg   = props.Color{Red: 30, Green: 30, Blue: 30}
	invHeaderBg  = props.Color{Red: 50, Green: 50, Blue: 50}
)

// ─── Table cell styles ─────────────────────────────────────────────────────────

// full border on every cell
func cellBorder(bg *props.Color) *props.Cell {
	return &props.Cell{
		BackgroundColor: bg,
		BorderColor:     &invBorder,
		BorderType:      border.Full,
		BorderThickness: 0.3,
	}
}

// header cell: dark bg, full border
func headerCell() *props.Cell {
	return &props.Cell{
		BackgroundColor: &invHeaderBg,
		BorderColor:     &invBorder,
		BorderType:      border.Full,
		BorderThickness: 0.3,
	}
}

// totals table cell: right-side only (cleaner look)
func totalsCell(bg *props.Color) *props.Cell {
	return &props.Cell{
		BackgroundColor: bg,
		BorderColor:     &invBorder,
		BorderType:      border.Full,
		BorderThickness: 0.3,
	}
}

// ─── GenerateInvoice ───────────────────────────────────────────────────────────

func GenerateInvoice(path string, d InvoiceData) error {
	cfg := config.NewBuilder().
		WithPageSize("A4").
		WithLeftMargin(15).WithRightMargin(15).
		WithTopMargin(15).WithBottomMargin(15).
		Build()

	m := maroto.New(cfg)

	// ── INVOICE title ─────────────────────────────────────────────────────────
	m.AddRows(
		row.New(18).Add(
			col.New(6).Add(text.New("INVOICE", props.Text{
				Size: 28, Style: fontstyle.Bold, Color: &invBlack,
			})),
			col.New(6).Add(text.New("STUDIO SHODWE", props.Text{
				Size: 10, Style: fontstyle.Bold, Color: &invDarkGray, Align: align.Right,
			})),
		),
	)

	m.AddRows(
		row.New(6).Add(col.New(12).Add(text.New(
			"Invoice Number: #"+d.Number+"     Invoice Date: "+d.Date.Format("02/01/2006"),
			props.Text{Size: 8, Color: &invDarkGray},
		))),
	)

	m.AddRows(row.New(5))

	// ── From / Bill To ────────────────────────────────────────────────────────
	m.AddRows(
		row.New(6).Add(
			col.New(6).Add(text.New(d.Provider.Name, props.Text{Size: 9, Style: fontstyle.Bold, Color: &invBlack})),
			col.New(6).Add(text.New("BILL TO", props.Text{Size: 9, Style: fontstyle.Bold, Color: &invBlack})),
		),
		row.New(5).Add(
			col.New(6).Add(text.New(d.Provider.Address+", "+d.Provider.City, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(6).Add(text.New(d.Client.Name, props.Text{Size: 8, Color: &invDarkGray})),
		),
		row.New(5).Add(
			col.New(6).Add(text.New(d.Provider.Phone, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(6).Add(text.New(d.Client.Address+", "+d.Client.City, props.Text{Size: 8, Color: &invDarkGray})),
		),
		row.New(5).Add(
			col.New(6).Add(text.New(d.Provider.Email, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(6).Add(text.New(d.Client.Phone, props.Text{Size: 8, Color: &invDarkGray})),
		),
		row.New(5).Add(
			col.New(6),
			col.New(6).Add(text.New(d.Client.Email, props.Text{Size: 8, Color: &invDarkGray})),
		),
	)

	m.AddRows(row.New(6))

	// ── Items Table ───────────────────────────────────────────────────────────
	// Header row
	m.AddRows(
		row.New(9).Add(
			col.New(6).WithStyle(headerCell()).Add(text.New("Item & Description", props.Text{Size: 9, Style: fontstyle.Bold, Color: &invWhite})),
			col.New(2).WithStyle(headerCell()).Add(text.New("Unit Price", props.Text{Size: 9, Style: fontstyle.Bold, Color: &invWhite, Align: align.Right})),
			col.New(2).WithStyle(headerCell()).Add(text.New("Qty", props.Text{Size: 9, Style: fontstyle.Bold, Color: &invWhite, Align: align.Center})),
			col.New(2).WithStyle(headerCell()).Add(text.New("Amount", props.Text{Size: 9, Style: fontstyle.Bold, Color: &invWhite, Align: align.Right})),
		),
	)

	// Data rows
	var itemRows []core.Row
	for i, it := range d.Items {
		it := it
		bg := &invWhite
		if i%2 != 0 {
			bg = &invLightGray
		}
		itemRows = append(itemRows, row.New(9).Add(
			col.New(6).WithStyle(cellBorder(bg)).Add(text.New(it.Description, props.Text{Size: 8, Color: &invBlack})),
			col.New(2).WithStyle(cellBorder(bg)).Add(text.New(fmt.Sprintf("$%.2f", it.UnitPrice), props.Text{Size: 8, Color: &invBlack, Align: align.Right})),
			col.New(2).WithStyle(cellBorder(bg)).Add(text.New(fmt.Sprintf("%d", it.Qty), props.Text{Size: 8, Color: &invBlack, Align: align.Center})),
			col.New(2).WithStyle(cellBorder(bg)).Add(text.New(fmt.Sprintf("$%.2f", it.Amount()), props.Text{Size: 8, Color: &invBlack, Align: align.Right})),
		))
	}
	m.AddRows(itemRows...)

	m.AddRows(row.New(6))

	// ── Notes + Totals ────────────────────────────────────────────────────────
	taxLabel := fmt.Sprintf("Tax (%.0f%%)", d.TaxRate*100)

	m.AddRows(
		row.New(7).Add(
			col.New(6).Add(text.New("NOTES / TERMS:", props.Text{Size: 8, Style: fontstyle.Bold, Color: &invBlack})),
			col.New(3),
			col.New(2).WithStyle(totalsCell(&invLightGray)).Add(text.New("Sub-Total", props.Text{Size: 8, Color: &invDarkGray})),
			col.New(1).WithStyle(totalsCell(&invLightGray)).Add(text.New(fmt.Sprintf("$%.2f", d.SubTotal()), props.Text{Size: 8, Align: align.Right, Color: &invBlack})),
		),
		row.New(7).Add(
			col.New(6).Add(text.New(d.Notes, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(3),
			col.New(2).WithStyle(totalsCell(&invWhite)).Add(text.New(taxLabel, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(1).WithStyle(totalsCell(&invWhite)).Add(text.New(fmt.Sprintf("$%.2f", d.Tax()), props.Text{Size: 8, Align: align.Right, Color: &invBlack})),
		),
		// Total row — dark background
		row.New(9).Add(
			col.New(9),
			col.New(2).WithStyle(&props.Cell{BackgroundColor: &invTotalBg, BorderType: border.Full, BorderColor: &invTotalBg, BorderThickness: 0.3}).Add(
				text.New("Total", props.Text{Size: 9, Style: fontstyle.Bold, Color: &invWhite}),
			),
			col.New(1).WithStyle(&props.Cell{BackgroundColor: &invTotalBg, BorderType: border.Full, BorderColor: &invTotalBg, BorderThickness: 0.3}).Add(
				text.New(fmt.Sprintf("$%.2f", d.Total()), props.Text{Size: 9, Style: fontstyle.Bold, Color: &invWhite, Align: align.Right}),
			),
		),
	)

	m.AddRows(row.New(10))

	// ── Footer divider ────────────────────────────────────────────────────────
	m.AddRows(row.New(2).Add(col.New(12).Add(line.New(props.Line{Color: &invBorder, Thickness: 0.5}))))
	m.AddRows(row.New(4))

	// ── Payment Method / Prepared By ─────────────────────────────────────────
	m.AddRows(
		row.New(6).Add(
			col.New(6).Add(text.New("PAYMENT METHOD", props.Text{Size: 8, Style: fontstyle.Bold, Color: &invBlack})),
			col.New(6).Add(text.New("PREPARED BY", props.Text{Size: 8, Style: fontstyle.Bold, Color: &invBlack})),
		),
		row.New(5).Add(
			col.New(6).Add(text.New("Bank: "+d.PaymentMethod.Bank, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(6).Add(text.New(d.PreparedBy.Name, props.Text{Size: 8, Color: &invDarkGray})),
		),
		row.New(5).Add(
			col.New(6).Add(text.New("Account Name: "+d.PaymentMethod.AccountName, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(6).Add(text.New(d.PreparedBy.Title, props.Text{Size: 8, Color: &invDarkGray})),
		),
		row.New(5).Add(
			col.New(6).Add(text.New("Account Number: "+d.PaymentMethod.AccountNumber, props.Text{Size: 8, Color: &invDarkGray})),
			col.New(6),
		),
	)

	doc, err := m.Generate()
	if err != nil {
		return err
	}
	return doc.Save(path)
}
