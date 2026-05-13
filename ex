package chromedp

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
	renderTimeout = 20 * time.Second
	renderSettle  = 2 * time.Second
)

func GenerateNamed(html, filename string, width, height float64) (string, error) {
	if err := os.MkdirAll("/temp", 0o755); err != nil {
		return "", err
	}
	outputPath := filepath.Join("/temp", filename)
	if err := renderToFile(html, outputPath, width, height); err != nil {
		return "", err
	}
	return outputPath, nil
}

func renderToFile(html, outputPath string, width, height float64) error {
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
				WithPaperWidth(width).
				WithPaperHeight(height).
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
