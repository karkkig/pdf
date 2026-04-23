package mapdf

import (
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"

	// "github.com/johnfercher/maroto/v2/pkg/components/image"
	// "github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
	qrcode "github.com/skip2/go-qrcode"
)

// ─── Data ──────────────────────────────────────────────────────────────────────

type GPayBadgeData struct {
	BusinessName string
	PhoneNumber  string
	UPIHandle    string // e.g. "12345 67890@yhh"
	// QRCodePath is the path to the QR code image file (PNG/JPG).
	// If empty, a placeholder box is drawn instead.
	QRContent string // e.g. "upi://pay?pa=1234567890@yhh&pn=Your+Business+Name&cu=INR"
	// QRCodePath string
}

// ─── GenerateGPayBadge ────────────────────────────────────────────────────────
// Produces a portrait card-sized PDF (85mm x 135mm) resembling the GPay badge.

func GenerateGPayBadge(path string, d GPayBadgeData) error {
	bgBlack := props.Color{Red: 0, Green: 0, Blue: 0}
	gpayBlue := props.Color{Red: 66, Green: 133, Blue: 244}
	gpayGray := props.Color{Red: 120, Green: 120, Blue: 120}
	white := props.Color{Red: 255, Green: 255, Blue: 255}

	qrBytes, err := qrcode.Encode(d.QRContent, qrcode.High, 512)
	if err != nil {
		return err
	}

	// Custom card size: 85mm wide x 135mm tall
	cfg := config.NewBuilder().
		WithDimensions(85, 135).
		WithLeftMargin(6).WithRightMargin(6).
		WithTopMargin(8).WithBottomMargin(8).
		Build()

	m := maroto.New(cfg)

	// ── GPay logo text ────────────────────────────────────────────────────────
	m.AddRows(
		row.New(12).Add(
			col.New(3),
			col.New(3).Add(
				text.New("G ", props.Text{
					Size:  20,
					Style: fontstyle.Bold,
					Color: &gpayBlue,
					Align: align.Right,
				}),
			),
			col.New(3).Add(
				text.New("Pay", props.Text{
					Size:  20,
					Style: fontstyle.Bold,
					Color: &bgBlack,
					Align: align.Left,
				}),
			),
			col.New(3),
		),
	)
	m.AddRows(
		row.New(6).Add(
			col.New(12).Add(text.New("LogoHere", props.Text{
				Size: 7, Color: &gpayGray, Align: align.Center,
			})),
		),
	)

	m.AddRows(row.New(3))

	// ── "accepted here" ───────────────────────────────────────────────────────
	m.AddRows(
		row.New(7).Add(
			col.New(12).Add(text.New("accepted here", props.Text{
				Size: 9, Color: &gpayGray, Align: align.Center,
			})),
		),
	)

	m.AddRows(row.New(4))
	m.AddRows(row.New(2).Add(col.New(12).Add(line.New(props.Line{Color: &gpayGray, Thickness: 0.3}))))
	m.AddRows(row.New(4))

	// ── Business name & phone ─────────────────────────────────────────────────
	m.AddRows(
		row.New(9).Add(
			col.New(12).Add(text.New(d.BusinessName, props.Text{
				Size: 11, Style: fontstyle.Bold,
				Color: &bgBlack, Align: align.Center,
			})),
		),
	)
	m.AddRows(
		row.New(7).Add(
			col.New(12).Add(text.New(d.PhoneNumber, props.Text{
				Size: 9, Color: &gpayGray, Align: align.Center,
			})),
		),
	)

	m.AddRows(row.New(2))

	m.AddRows(
		row.New(42).Add(
			col.New(1),
			col.New(10).Add(
				image.NewFromBytes(qrBytes, extension.Png, props.Rect{
					Center:  true,
					Percent: 95,
				}),
			),
			col.New(1),
		),
	)

	// m.AddRows(
	// 	row.New(42).Add(
	// 		col.New(1),
	// 		code.NewQrCol(10, d.QRContent, props.Rect{
	// 			Center:  true,
	// 			Percent: 95,
	// 		}),
	// 		col.New(1),
	// 	),
	// )

	m.AddRows(row.New(2))

	//m.AddRows(row.New(4))

	// ── QR Code placeholder (replace with image.NewFromFileCol when you have the file) ──
	// qrSize := 32.0
	// if d.QRCodePath != "" {
	// 	m.AddRows(
	// 		row.New(qrSize).Add(
	// 			col.New(2),
	// 			image.NewFromFileCol(
	// 				8,
	// 				d.QRCodePath,
	// 				props.Rect{
	// 					Center:  true,
	// 					Percent: 100,
	// 				},
	// 			),
	// 			col.New(2),
	// 		),
	// 	)
	// } else {
	// 	// fallback placeholder (optional)
	// 	qrBg := props.Color{Red: 230, Green: 230, Blue: 230}
	// 	m.AddRows(
	// 		row.New(qrSize).WithStyle(&props.Cell{BackgroundColor: &qrBg}).Add(
	// 			col.New(2),
	// 			col.New(8).Add(text.New("[ QR CODE ]", props.Text{
	// 				Size: 9, Color: &gpayGray, Align: align.Center,
	// 			})),
	// 			col.New(2),
	// 		),
	// 	)
	// }

	// m.AddRows(row.New(4))

	// ── UPI handle ───────────────────────────────────────────────────────────
	m.AddRows(
		row.New(7).Add(
			col.New(12).Add(text.New(d.UPIHandle, props.Text{
				Size: 9, Style: fontstyle.Bold,
				Color: &bgBlack, Align: align.Center,
			})),
		),
	)

	// Bottom accent bar
	m.AddRows(row.New(3))
	m.AddRows(row.New(3).WithStyle(&props.Cell{BackgroundColor: &gpayBlue}).Add(col.New(12)))

	_ = white // available for further styling

	doc, err := m.Generate()
	if err != nil {
		return err
	}
	return doc.Save(path)
}
