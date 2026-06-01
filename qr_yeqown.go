package qrgen

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// CreateQRWithLogo generates a QR code PNG using the yeqown library.
//
//   - ctx       – request context; cancellation aborts logo fetching
//   - content   – data to encode
//   - logoURL   – optional URL of the logo to embed in the centre
//   - dimension – output image size in pixels
//   - border    – quiet-zone border width in pixels
//   - cfg       – service-level configuration (size limits, logo ratio, etc.)
//
// Returns the path to a temporary PNG file. The caller must remove it.
func CreateQRWithLogo(ctx context.Context, content, logoURL string, dimension, border int, cfg Config) (string, error) {
	slog.Info("creating QR code (yeqown)", "dimension", dimension, "border", border, "hasLogo", logoURL != "")

	qr, err := qrcode.New(content)
	if err != nil {
		return "", fmt.Errorf("create qrcode: %w", err)
	}

	version := qrVersionFromContent(content)
	modules := (version-1)*4 + 21
	qrWidth := uint8(dimension / modules)
	if qrWidth < 1 {
		qrWidth = 1
	}

	slog.Info("QR parameters", "version", version, "modules", modules, "qrWidthPerModule", qrWidth)

	// Reserve a temp output path; clean up on any error via success flag.
	tmpOut, err := os.CreateTemp("", "qrcode-*.png")
	if err != nil {
		return "", fmt.Errorf("create temp output file: %w", err)
	}
	outPath := tmpOut.Name()
	tmpOut.Close() // yeqown opens by path; release our handle first

	success := false
	defer func() {
		if !success {
			os.Remove(outPath)
		}
	}()

	var options []standard.ImageOption

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

		nativeQRSize := int(qrWidth) * modules
		targetLogoSize := int(float64(nativeQRSize) * cfg.LogoSizeRatio)
		slog.Info("logo sizing", "nativeQRSize", nativeQRSize, "targetLogoSize", targetLogoSize)

		if err = resizeLogoToTarget(logoPath, targetLogoSize); err != nil {
			return "", fmt.Errorf("resize logo: %w", err)
		}

		options = []standard.ImageOption{
			standard.WithLogoImageFileJPEG(logoPath),
			standard.WithQRWidth(qrWidth),
			standard.WithBorderWidth(border),
		}
	} else {
		options = []standard.ImageOption{
			standard.WithQRWidth(qrWidth),
			standard.WithBorderWidth(border),
		}
	}

	writer, err := standard.New(outPath, options...)
	if err != nil {
		return "", fmt.Errorf("create writer: %w", err)
	}
	defer writer.Close()

	if err = qr.Save(writer); err != nil {
		return "", fmt.Errorf("save qrcode: %w", err)
	}

	if err = resizeImage(outPath, dimension); err != nil {
		return "", fmt.Errorf("resize output: %w", err)
	}

	success = true
	slog.Info("QR code ready (yeqown)", "path", outPath)
	return outPath, nil
}
