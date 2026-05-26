package chromedp

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"os"

	qrgode "github.com/ahmedtahas/qr-gode"
)

// CreateQRWithLogo generates a QR code using the ahmedtahas/qr-gode library.
// Outputs a PNG file named "qrcode_with_logo.png" with a white background.
func CreateQRWithLogo(content string, logoURL string, dimension int) error {
	fmt.Printf("Creating QR code with content: %s, logoURL: %s, dimension: %d\n",
		content, logoURL, dimension)

	builder := qrgode.New(content).
		Size(dimension).
		ErrorCorrection(qrgode.LevelH)

	if logoURL != "" {
		if err := UrlGet(logoURL); err != nil {
			fmt.Printf("failed to fetch logo: %v\n", err)
			return err
		}

		builder = builder.Logo("logo1.jpg")

		for _, w := range builder.ScannabilityWarnings() {
			fmt.Printf("scannability warning: %s\n", w)
		}
	}

	// Get PNG bytes from the library
	pngBytes, err := builder.PNG()
	if err != nil {
		fmt.Printf("generate qrcode failed: %v\n", err)
		return err
	}

	// Decode the generated PNG
	qrImg, err := png.Decode(bytesReader(pngBytes))
	if err != nil {
		fmt.Printf("decode qrcode failed: %v\n", err)
		return err
	}

	// Composite onto a white canvas
	bounds := qrImg.Bounds()
	white := image.NewRGBA(bounds)
	draw.Draw(white, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(white, bounds, qrImg, image.Point{}, draw.Over)

	// Save to file
	out, err := os.Create("qrcode_with_logo.png")
	if err != nil {
		return err
	}
	defer out.Close()

	if err = png.Encode(out, white); err != nil {
		fmt.Printf("save qrcode failed: %v\n", err)
		return err
	}

	fmt.Println("QR code saved to qrcode_with_logo.png")
	return nil
}

// bytesReader wraps a byte slice as an io.Reader.
func bytesReader(b []byte) io.Reader {
	return &bytesReaderImpl{b: b, pos: 0}
}

type bytesReaderImpl struct {
	b   []byte
	pos int
}

func (r *bytesReaderImpl) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.b) {
		return 0, io.EOF
	}
	n = copy(p, r.b[r.pos:])
	r.pos += n
	return n, nil
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
