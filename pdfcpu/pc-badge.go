package pcpdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"image/color"
	_ "image/png" // register PNG decoder
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	qrcode "github.com/skip2/go-qrcode"
)

// GPayBadgeData holds all fields for the badge.
type GPayBadgeData struct {
	BusinessName string
	PhoneNumber  string
	QRContent    string // UPI deep-link: "upi://pay?pa=handle@yhh&pn=Name&cu=INR"
	UPIHandle    string // shown below QR
}

// pdfEscape escapes special characters in PDF literal strings

// ascii85Encode encodes bytes to ASCII85 for use with /ASCII85Decode filter.
func ascii85Encode(src []byte) string {
	var buf bytes.Buffer
	for i := 0; i < len(src); i += 4 {
		end := i + 4
		if end > len(src) {
			end = len(src)
		}
		n := end - i
		var b uint32
		for j := 0; j < 4; j++ {
			b <<= 8
			if j < n {
				b |= uint32(src[i+j])
			}
		}
		if n == 4 && b == 0 {
			buf.WriteByte('z')
			continue
		}
		var out [5]byte
		for j := 4; j >= 0; j-- {
			out[j] = byte(b%85) + '!'
			b /= 85
		}
		buf.Write(out[:n+1])
	}
	buf.WriteString("~>")
	return buf.String()
}

// pngToZlibRGB decodes a PNG image and returns (zlibCompressedRGBPixels, width, height, err).
// PDF /FlateDecode expects raw scanline pixels with a filter byte prefix per row.
// We use PNG-style filter byte 0x00 (None) per row for simplicity.
func pngToZlibRGB(pngBytes []byte) ([]byte, int, int, error) {
	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, 0, 0, err
	}
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	// Build raw scanlines: each row prefixed with filter byte 0x00
	var raw bytes.Buffer
	for y := 0; y < h; y++ {
		raw.WriteByte(0x00) // PNG filter type None
		for x := 0; x < w; x++ {
			c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			raw.WriteByte(c.R)
			raw.WriteByte(c.G)
			raw.WriteByte(c.B)
		}
	}

	// zlib-compress the scanlines
	var zlibBuf bytes.Buffer
	zw := zlib.NewWriter(&zlibBuf)
	zw.Write(raw.Bytes())
	zw.Close()

	return zlibBuf.Bytes(), w, h, nil
}

// GenerateGPayBadge writes a crisp GPay-style payment badge PDF using pdfcpu.
//
// Dependencies:
//
//	go get github.com/skip2/go-qrcode
//	go get github.com/pdfcpu/pdfcpu
func GenerateGPayBadge(outPath string, d GPayBadgeData) error {
	// ── 1. Generate QR at 512px — balanced between sharp and compact ──────────
	qrPNG, err := qrcode.Encode(d.QRContent, qrcode.High, 512)
	if err != nil {
		return fmt.Errorf("qr encode: %w", err)
	}

	// Decode PNG → zlib-compressed raw RGB pixels for embedding in PDF
	qrPixels, qrW, qrH, err := pngToZlibRGB(qrPNG)
	if err != nil {
		return fmt.Errorf("png decode: %w", err)
	}
	qrStream := ascii85Encode(qrPixels)

	// ── 2. Page dimensions (points; 72pt = 1 inch) ────────────────────────────
	// 200 × 320 pt = ~70 × 113 mm — compact portrait card matching the image
	W, H := 200.0, 320.0

	// ── 3. Content stream ─────────────────────────────────────────────────────
	var cs bytes.Buffer
	wf := func(f string, a ...any) { fmt.Fprintf(&cs, f, a...) }

	fillRect := func(x, y, rw, rh, r, g, b float64) {
		wf("%.3f %.3f %.3f rg %.2f %.2f %.2f %.2f re f\n", r, g, b, x, y, rw, rh)
	}
	hline := func(x1, x2, y, lw, r, g, b float64) {
		wf("%.3f %.3f %.3f RG %.2f w %.2f %.2f m %.2f %.2f l S\n", r, g, b, lw, x1, y, x2, y)
	}
	putText := func(font string, size, x, y, r, g, b float64, s string) {
		wf("BT /%s %.2f Tf %.3f %.3f %.3f rg %.2f %.2f Td (%s) Tj ET\n",
			font, size, r, g, b, x, y, pdfEscape(s))
	}
	// approximate centering for Helvetica (char width ≈ size*0.45)
	center := func(font string, size, y, r, g, b float64, s string) {
		x := (W - float64(len(s))*size*0.45) / 2
		putText(font, size, x, y, r, g, b, s)
	}

	// White background
	fillRect(0, 0, W, H, 1, 1, 1)

	// "G" blue + "Pay" dark
	logoY := H - 42.0
	putText("Helvetica-Bold", 22, W/2-24, logoY, 0.259, 0.522, 0.957, "G")
	putText("Helvetica-Bold", 18, W/2-2, logoY, 0.13, 0.13, 0.13, "Pay")

	center("Helvetica", 7, H-55, 0.65, 0.65, 0.65, "LogoHere")
	center("Helvetica", 8, H-68, 0.50, 0.50, 0.50, "accepted here")

	hline(18, W-18, H-78, 0.4, 0.82, 0.82, 0.82)

	center("Helvetica-Bold", 11, H-98, 0.08, 0.08, 0.08, d.BusinessName)
	center("Helvetica", 9, H-113, 0.38, 0.38, 0.38, d.PhoneNumber)

	// QR image — placed via XObject /QR
	qrSize := 148.0
	qrX := (W - qrSize) / 2
	qrY := 68.0

	// white bg + light border behind QR
	fillRect(qrX-3, qrY-3, qrSize+6, qrSize+6, 1, 1, 1)
	wf("0.85 0.85 0.85 RG 0.5 w %.2f %.2f %.2f %.2f re S\n", qrX-3, qrY-3, qrSize+6, qrSize+6)

	// draw the QR XObject
	wf("q %.2f 0 0 %.2f %.2f %.2f cm /QR Do Q\n", qrSize, qrSize, qrX, qrY)

	center("Helvetica-Bold", 9, qrY-18, 0.08, 0.08, 0.08, d.UPIHandle)

	// Google-colour bottom bar (6 equal segments)
	segW := W / 6
	barColors := [][3]float64{
		{0.259, 0.522, 0.957}, // blue
		{0.918, 0.263, 0.208}, // red
		{0.984, 0.737, 0.016}, // yellow
		{0.259, 0.522, 0.957}, // blue
		{0.204, 0.659, 0.325}, // green
		{0.918, 0.263, 0.208}, // red
	}
	for i, c := range barColors {
		fillRect(float64(i)*segW, 0, segW, 14, c[0], c[1], c[2])
	}

	pageStream := cs.String()

	// ── 4. Assemble raw PDF ───────────────────────────────────────────────────
	var raw bytes.Buffer
	off := make([]int, 7) // track byte offsets for xref, objects 1–6 used

	wr := func(s string) { raw.WriteString(s) }
	wrl := func(s string) { raw.WriteString(s + "\n") }
	wrf := func(f string, a ...any) { fmt.Fprintf(&raw, f, a...) }

	wrl("%PDF-1.4")
	wrl("%\xe2\xe3\xcf\xd3") // binary marker

	// obj 1: Catalog
	off[1] = raw.Len()
	wrl("1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj")

	// obj 2: Pages
	off[2] = raw.Len()
	wrl("2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj")

	// obj 3: Page — references content (4) and QR image (5)
	off[3] = raw.Len()
	wrl("3 0 obj")
	wrf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f]\n", W, H)
	wrl("   /Contents 4 0 R")
	wrl("   /Resources <<")
	wrl("     /Font <<")
	wrl("       /Helvetica      << /Type /Font /Subtype /Type1 /BaseFont /Helvetica      /Encoding /WinAnsiEncoding >>")
	wrl("       /Helvetica-Bold << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")
	wrl("     >>")
	wrl("     /XObject << /QR 5 0 R >>")
	wrl("   >>")
	wrl(">> endobj")

	// obj 4: content stream
	off[4] = raw.Len()
	wrl("4 0 obj")
	wrf("<< /Length %d >>\n", len(pageStream))
	wrl("stream")
	wr(pageStream)
	wrl("\nendstream endobj")

	// obj 5: QR image XObject
	// Filter chain: ASCII85Decode → FlateDecode → raw RGB scanlines with filter bytes
	// /DecodeParms for FlateDecode: Predictor 15 = PNG adaptive, Colors 3, Columns = qrW
	off[5] = raw.Len()
	wrl("5 0 obj")
	wrf("<< /Type /XObject /Subtype /Image /Width %d /Height %d\n", qrW, qrH)
	wrl("   /ColorSpace /DeviceRGB /BitsPerComponent 8")
	wrl("   /Filter [/ASCII85Decode /FlateDecode]")
	wrf("   /DecodeParms [null << /Predictor 15 /Colors 3 /BitsPerComponent 8 /Columns %d >>]\n", qrW)
	wrf("   /Length %d >>\n", len(qrStream))
	wrl("stream")
	wr(qrStream)
	wrl("\nendstream endobj")

	// xref
	xrefPos := raw.Len()
	wrl("xref")
	wrf("0 6\n")
	wrl("0000000000 65535 f ")
	for i := 1; i <= 5; i++ {
		wrf("%010d 00000 n \n", off[i])
	}
	wrl("trailer << /Size 6 /Root 1 0 R >>")
	wrl("startxref")
	wrf("%d\n", xrefPos)
	wr("%%EOF")

	// ── 5. Run through pdfcpu to normalize + validate ─────────────────────────
	rs := bytes.NewReader(raw.Bytes())
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return api.Optimize(rs, out, conf)
}
