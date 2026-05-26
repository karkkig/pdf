package chromedp

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/nfnt/resize"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// CreateQRWithLogo generates a QR code using the WithLogo option.
// The QR width is calculated based on the logo size and actual QR version,
// then the final image is resized to dimensionxdimension.
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

		qrWidth, err := qrWidthFromLogo("logo1.jpg", qr)
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

	if err = resizeImage("qrcode_with_logo.png", dimension); err != nil {
		fmt.Printf("resize failed: %v\n", err)
		return err
	}

	return nil
}

// qrWidthFromLogo reads the logo file and calculates an appropriate QR module
// width so the logo occupies roughly 1/5 (20%) of the total QR side length,
// based on the actual QR version's module count.
//
// modules = (version-1)*4 + 21
// totalSide = logoSide * 5
// qrWidth = totalSide / modules
func qrWidthFromLogo(logoPath string, qr *qrcode.QRCode) (uint8, error) {
	f, err := os.Open(logoPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, fmt.Errorf("decode logo config: %w", err)
	}

	logoSide := cfg.Width
	if cfg.Height > logoSide {
		logoSide = cfg.Height
	}

	// Derive actual module count from the QR version
	version := qr.Version()
	modules := (version-1)*4 + 21

	// totalSide = logoSide * 5  →  qrWidth = totalSide / modules
	qrWidth := int(float64(logoSide*5) / float64(modules))
	if qrWidth < 1 {
		qrWidth = 1
	}
	if qrWidth > 255 {
		qrWidth = 255
	}

	fmt.Printf("Logo side: %dpx, QR version: %d, modules: %d → QR module width: %d (total QR ~%dpx)\n",
		logoSide, version, modules, qrWidth, qrWidth*modules)

	return uint8(qrWidth), nil
}

// resizeImage reads any supported image at path, resizes it to size×size,
// and overwrites the file as PNG.
func resizeImage(path string, size int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	img, format, err := image.Decode(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("decode image (format=%s): %w", format, err)
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
