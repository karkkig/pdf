package qrgen

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log/slog"
	"net/http"
	"os"

	qrgode "github.com/ahmedtahas/qr-gode"
)

// CreateQRWithLogo2 generates a QR code PNG using the qr-gode library.
//
//   - ctx       – request context; cancellation aborts logo fetching
//   - content   – data to encode
//   - logoURL   – optional URL of the logo to embed in the centre
//   - dimension – output image size in pixels
//   - border    – quiet-zone width in modules
//   - cfg       – service-level configuration (size limits, etc.)
//
// Returns the path to a temporary PNG file. The caller must remove it.
func CreateQRWithLogo2(ctx context.Context, content, logoURL string, dimension, border int, cfg Config) (string, error) {
	slog.Info("creating QR code (qr-gode)", "dimension", dimension, "border", border, "hasLogo", logoURL != "")

	builder := qrgode.New(content).
		Size(dimension).
		QuietZone(border).
		LogoBackground("transparent").
		LogoMode(qrgode.LogoOverlay).
		ErrorCorrection(qrgode.LevelH)

	if logoURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, logoURL, nil)
		if err != nil {
			return "", fmt.Errorf("build logo request: %w", err)
		}
		logoPath, err := UrlGet(req, cfg.MaxLogoBytes)
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

	pngBytes, err := builder.PNG()
	if err != nil {
		return "", fmt.Errorf("generate qr png: %w", err)
	}

	qrImg, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return "", fmt.Errorf("decode qr png: %w", err)
	}

	// Composite onto a white background so any transparency becomes white.
	bounds := qrImg.Bounds()
	white := image.NewRGBA(bounds)
	draw.Draw(white, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(white, bounds, qrImg, image.Point{}, draw.Over)

	// Write to a temp file; clean up on any error via success flag.
	out, err := os.CreateTemp("", "qrcode-*.png")
	if err != nil {
		return "", fmt.Errorf("create temp output file: %w", err)
	}
	outPath := out.Name()

	success := false
	defer func() {
		if !success {
			os.Remove(outPath)
		}
	}()

	if err = png.Encode(out, white); err != nil {
		out.Close()
		return "", fmt.Errorf("encode qr png: %w", err)
	}
	out.Close()

	success = true
	slog.Info("QR code ready (qr-gode)", "path", outPath)
	return outPath, nil
}
