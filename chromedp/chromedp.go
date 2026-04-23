package cdpdf

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func ReadHTML(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func GeneratePDF(html string, output string) error {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var pdfBuf []byte

	formathtml := `
	<html>
	<head>
		<meta charset="UTF-8">
		<style>
			body { margin: 0; }
		</style>
	</head>
	<body>
		` + html + `
	</body>
	</html>`

	htmlURL := "data:text/html," + url.PathEscape(formathtml)

	err := chromedp.Run(ctx,
		chromedp.Navigate(htmlURL),
		chromedp.Sleep(2*time.Second), // wait for rendering

		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).   // A4 width
				WithPaperHeight(11.69). // A4 height
				Do(ctx)

			pdfBuf = buf
			return err
		}),
	)

	if err != nil {
		return err
	}

	return os.WriteFile(output, pdfBuf, 0644)
}
