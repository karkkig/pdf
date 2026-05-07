package pdf

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	paperWidthIn  = 8.27  // A4
	paperHeightIn = 11.69 // A4
	renderTimeout = 20 * time.Second
	renderSettle  = 2 * time.Second
)

// Generate renders html to a PDF and writes it to a temporary file.
// The caller is responsible for removing the file after use.
func Generate(html string) (path string, err error) {
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

// GenerateNamed is like Generate but uses a caller-supplied filename stem
// inside the OS temp directory. Useful when the response filename matters.
func GenerateNamed(html, filename string) (path string, err error) {
	outputPath := filepath.Join(os.TempDir(), filename)

	if err = renderToFile(html, outputPath); err != nil {
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
