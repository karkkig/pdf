package qrgen

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// CreateQRWithLogo generates a QR code using the yeqown library.
//
//   - content   – data to encode in the QR code
//   - logoURL   – optional URL of the logo image to embed in the center
//   - dimension – pixel dimensions of the output image (e.g. 300 for 300×300)
//   - border    – quiet-zone border width in pixels around the QR code
//
// Returns the path to the generated PNG file or an error.
func CreateQRWithLogo(content string, logoURL string, dimension int, border int) (string, error) {
	slog.Info("creating QR code", "content", content, "logoURL", logoURL, "dimension", dimension, "border", border)

	qr, err := qrcode.New(content)
	if err != nil {
		slog.Error("create qrcode failed", "error", err)
		return "", fmt.Errorf("create qrcode: %w", err)
	}

	version := qrVersionFromContent(content)
	modules := (version-1)*4 + 21
	qrWidth := uint8(dimension / modules)
	if qrWidth < 1 {
		qrWidth = 1
	}

	slog.Info("QR parameters", "version", version, "modules", modules, "qrWidthPerModule", qrWidth)

	// Write to a temp file to avoid races between concurrent requests.
	tmpOut, err := os.CreateTemp("", "qrcode-*.png")
	if err != nil {
		return "", fmt.Errorf("create temp output file: %w", err)
	}
	outPath := tmpOut.Name()
	tmpOut.Close() // yeqown opens the file by path; close our handle first.

	var options []standard.ImageOption

	if logoURL != "" {
		logoPath, err := UrlGet(logoURL)
		if err != nil {
			slog.Error("failed to fetch logo", "url", logoURL, "error", err)
			os.Remove(outPath)
			return "", fmt.Errorf("fetch logo: %w", err)
		}
		defer os.Remove(logoPath)

		nativeQRSize := int(qrWidth) * modules
		targetLogoSize := nativeQRSize / 5

		slog.Info("logo sizing", "nativeQRSize", nativeQRSize, "targetLogoSize", targetLogoSize)

		if err = resizeLogoToTarget(logoPath, targetLogoSize); err != nil {
			slog.Error("failed to resize logo", "error", err)
			os.Remove(outPath)
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
		slog.Error("create writer failed", "error", err)
		os.Remove(outPath)
		return "", fmt.Errorf("create writer: %w", err)
	}
	defer writer.Close()

	if err = qr.Save(writer); err != nil {
		slog.Error("save qrcode failed", "error", err)
		os.Remove(outPath)
		return "", fmt.Errorf("save qrcode: %w", err)
	}

	if err = resizeImage(outPath, dimension); err != nil {
		slog.Error("resize output failed", "error", err)
		os.Remove(outPath)
		return "", fmt.Errorf("resize output: %w", err)
	}

	slog.Info("QR code saved", "path", outPath)
	return outPath, nil
}
