package pdf

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ─── Constants ────────────────────────────────────────────────────────────────

const (
	paperWidthIn  = 8.27  // A4
	paperHeightIn = 11.69 // A4
	renderTimeout = 20 * time.Second
	renderSettle  = 2 * time.Second
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrMissingFilename = errors.New("X-PDF-Name header is required")
	ErrEmptyBody       = errors.New("HTML body must not be empty")
)

// ─── Validation ───────────────────────────────────────────────────────────────

// SanitizeFilename cleans the provided name, ensures a .pdf extension,
// and strips any path components to prevent traversal.
func SanitizeFilename(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", ErrMissingFilename
	}

	name = filepath.Base(name)
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		name += ".pdf"
	}

	return name, nil
}

// ValidateBody returns ErrEmptyBody when the HTML string is blank.
func ValidateBody(html string) error {
	if strings.TrimSpace(html) == "" {
		return ErrEmptyBody
	}
	return nil
}

// ─── Generation ───────────────────────────────────────────────────────────────

// Generate renders html to a PDF and writes it to a temporary file.
// The caller is responsible for removing the file after use.
func Generate(html string) (string, error) {
	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		return "", err
	}
	tmpFile.Close()

	outputPath := tmpFile.Name()
	if err = renderToFile(html, outputPath); err != nil {
		os.Remove(outputPath)
		return "", err
	}

	return outputPath, nil
}

// GenerateNamed is like Generate but places the file in the OS temp
// directory under the given filename. Useful when the download name matters.
func GenerateNamed(html, filename string) (string, error) {
	outputPath := filepath.Join(os.TempDir(), filename)
	if err := renderToFile(html, outputPath); err != nil {
		return "", err
	}
	return outputPath, nil
}

// renderToFile is the single place that talks to chromedp.
func renderToFile(html, outputPath string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, renderTimeout)
	defer cancel()

	wrapped := `<html><head><meta charset="UTF-8">` +
		`<style>body{margin:0;}</style></head><body>` + html + `</body></html>`

	htmlURL := "data:text/html," + url.PathEscape(wrapped)

	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(htmlURL),
		chromedp.Sleep(renderSettle),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(paperWidthIn).
				WithPaperHeight(paperHeightIn).
				Do(ctx)
			pdfBuf = buf
			return err
		}),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdfBuf, 0o644)
}
