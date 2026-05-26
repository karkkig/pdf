package chromedp

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/nfnt/resize"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// CreateQRWithLogo generates a QR code using the WithLogo option.
// The QR width is calculated based on the logo size, then the final image is resized to dimensionxdimension.
func CreateQRWithLogo(content string, logoURL string, dimension int) error {
	fmt.Printf("Creating QR code with content: %s, logoURL: %s, dimension: %d\n",
		content, logoURL, dimension)

	qr, err := qrcode.New(content)
	if err != nil {
		fmt.Printf("create qrcode failed: %v\n", err)
		return err
	}

	var options []standard.ImageOption

	if logoURL != "" {
		if err = UrlGet(logoURL); err != nil {
			fmt.Printf("failed to fetch logo: %v\n", err)
			return err
		}

		// Derive QR module width from logo's actual pixel dimensions
		qrWidth, err := qrWidthFromLogo("logo1.jpg")
		if err != nil {
			fmt.Printf("failed to derive QR width from logo: %v\n", err)
			return err
		}
		fmt.Printf("Derived QR width from logo: %d\n", qrWidth)

		options = []standard.ImageOption{
			standard.WithLogoImageFileJPEG("logo1.jpg"),
			standard.WithQRWidth(qrWidth),
			standard.WithBorderWidth(0),
		}
	} else {
		// Fallback: no logo, use a sensible default module size
		options = []standard.ImageOption{
			standard.WithQRWidth(10),
			standard.WithBorderWidth(0),
		}
	}

	writer, err := standard.New("qrcode_with_logo.png", options...)
	if err != nil {
		fmt.Printf("create writer failed: %v\n", err)
		return err
	}
	defer writer.Close()

	if err = qr.Save(writer); err != nil {
		fmt.Printf("save qrcode failed: %v\n", err)
		return err
	}
	writer.Close() // flush before we read the file back

	// Resize the generated QR image to exactly dimensionxdimension
	if err = resizeImage("qrcode_with_logo.png", dimension); err != nil {
		fmt.Printf("resize failed: %v\n", err)
		return err
	}

	return nil
}

// qrWidthFromLogo reads the logo file, calculates an appropriate QR module
// width so the logo occupies roughly 20% of the total QR area.
//
// QR total side ≈ 21 modules × qrWidth  (version-1 baseline)
// Logo should fill ~20% → logoSide ≈ 0.20 × totalSide
//   → qrWidth = logoSide / (21 × 0.20) = logoSide / 4.2
func qrWidthFromLogo(logoPath string) (uint8, error) {
	f, err := os.Open(logoPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, fmt.Errorf("decode logo config: %w", err)
	}

	// Use the larger dimension so portrait/landscape logos both fit
	logoSide := cfg.Width
	if cfg.Height > logoSide {
		logoSide = cfg.Height
	}

	// 21 modules * 0.20 logo ratio = 4.2 → round to nearest int, clamp to [1,255]
	qrWidth := int(float64(logoSide) / 4.2)
	if qrWidth < 1 {
		qrWidth = 1
	}
	if qrWidth > 255 {
		qrWidth = 255
	}
	return uint8(qrWidth), nil
}

// resizeImage reads a PNG at path, resizes it to size×size, and overwrites the file.
func resizeImage(path string, size int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("decode png: %w", err)
	}

	resized := resize.Resize(uint(size), uint(size), img, resize.Lanczos3)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, resized)
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
