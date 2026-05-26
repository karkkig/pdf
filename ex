package chromedp

import (
	"fmt"
	"io"
	"net/http"
	"os"

	qrgode "github.com/ahmedtahas/qr-gode"
)

// CreateQRWithLogo generates a QR code using the ahmedtahas/qr-gode library.
// Outputs a PNG file named "qrcode_with_logo.png".
// The logo (if provided) is auto-sized to 15-30% of the QR canvas by the library.
// The border parameter maps to the quiet zone (margin) in modules.
func CreateQRWithLogo(content string, logoURL string, dimension int, border uint) error {
	fmt.Printf("Creating QR code with content: %s, logoURL: %s, dimension: %d, border: %d\n",
		content, logoURL, dimension, border)

	builder := qrgode.New(content).
		Size(dimension).
		QuietZone(int(border)).
		ErrorCorrection(qrgode.LevelH) // LevelH recommended when a logo is present

	if logoURL != "" {
		if err := UrlGet(logoURL); err != nil {
			fmt.Printf("failed to fetch logo: %v\n", err)
			return err
		}

		// Warn if the logo + ECL combination risks unscannability
		for _, w := range builder.Logo("logo1.jpg").ScannabilityWarnings() {
			fmt.Printf("scannability warning: %s\n", w)
		}

		builder = builder.Logo("logo1.jpg")
	}

	if err := builder.SaveAs("qrcode_with_logo.png"); err != nil {
		fmt.Printf("save qrcode failed: %v\n", err)
		return err
	}

	fmt.Println("QR code saved to qrcode_with_logo.png")
	return nil
}

// UrlGet downloads the resource at url and saves it as logo1.jpg.
func UrlGet(url string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	outfile, err := os.Create("logo1.jpg")
	if err != nil {
		return err
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, resp.Body)
	return err
}
