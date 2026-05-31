package qrgen

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log/slog"
	"os"

	qrgode "github.com/ahmedtahas/qr-gode"
)

// CreateQRWithLogo2 generates a QR code using the qr-gode library.
//
//   - content   – data to encode in the QR code
//   - logoURL   – optional URL of the logo image to embed in the center
//   - dimension – pixel dimensions of the output image (e.g. 300 for 300×300)
//   - border    – quiet-zone width in modules around the QR code
//
// Returns the path to the generated PNG file or an error.
func CreateQRWithLogo2(content string, logoURL string, dimension int, border int) (string, error) {
	builder := qrgode.New(content).
		Size(dimension).
		QuietZone(border).
		LogoBackground("transparent").
		LogoMode(qrgode.LogoOverlay).
		ErrorCorrection(qrgode.LevelH)

	if logoURL != "" {
		logoPath, err := UrlGet(logoURL)
		if err != nil {
			slog.Error("failed to fetch logo", "url", logoURL, "error", err)
			return "", fmt.Errorf("fetch logo: %w", err)
		}
		defer os.Remove(logoPath)

		builder = builder.Logo(logoPath)

		for _, w := range builder.ScannabilityWarnings() {
			slog.Warn("QR scannability warning", "warning", w)
		}
	}

	// Get PNG bytes from the library.
	pngBytes, err := builder.PNG()
	if err != nil {
		slog.Error("generate qrcode failed", "error", err)
		return "", fmt.Errorf("generate qr png: %w", err)
	}

	// Decode the generated PNG.
	qrImg, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		slog.Error("decode qrcode failed", "error", err)
		return "", fmt.Errorf("decode qr png: %w", err)
	}

	// Composite onto a white background so transparency becomes white.
	bounds := qrImg.Bounds()
	white := image.NewRGBA(bounds)
	draw.Draw(white, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(white, bounds, qrImg, image.Point{}, draw.Over)

	// Write to a temp file to avoid races between concurrent requests.
	out, err := os.CreateTemp("", "qrcode-*.png")
	if err != nil {
		return "", fmt.Errorf("create temp output file: %w", err)
	}
	outPath := out.Name()

	if err = png.Encode(out, white); err != nil {
		out.Close()
		os.Remove(outPath)
		slog.Error("save qrcode failed", "error", err)
		return "", fmt.Errorf("encode qr png: %w", err)
	}
	out.Close()

	slog.Info("QR code saved", "path", outPath)
	return outPath, nil
}
