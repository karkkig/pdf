package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/skip2/go-qrcode"
	xdraw "golang.org/x/image/draw"
)

func main() {
	content  := flag.String("content",   "https://example.com", "Data to encode in the QR code")
	size     := flag.Int("size",         512,                   "Width/height of the QR image in pixels")
	logoURL  := flag.String("logo",      "",                    "URL of the logo image (PNG or JPEG)")
	fgHex    := flag.String("fg",        "#000000",             "Foreground (module) colour as #RRGGBB")
	bgHex    := flag.String("bg",        "#FFFFFF",             "Background colour as #RRGGBB")
	output   := flag.String("output",    "qr_output.png",       "Output file path (.png)")
	logoFrac := flag.Float64("logofrac", 0.25,                  "Logo size as a fraction of QR size (0.0-0.35)")
	flag.Parse()

	if *logoFrac < 0 || *logoFrac > 0.35 {
		fmt.Fprintln(os.Stderr, "logofrac must be between 0.0 and 0.35")
		os.Exit(1)
	}

	fg, err := parseHex(*fgHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid fg colour: %v\n", err)
		os.Exit(1)
	}
	bg, err := parseHex(*bgHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid bg colour: %v\n", err)
		os.Exit(1)
	}

	// Use High error correction so the logo can cover ~30% without breaking reads.
	qr, err := qrcode.New(*content, qrcode.High)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qr generation failed: %v\n", err)
		os.Exit(1)
	}

	rawImg := qr.Image(*size)

	coloured := recolour(rawImg, fg, bg, *size)

	if *logoURL != "" {
		logoImg, err := fetchImage(*logoURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not fetch logo: %v\n", err)
			os.Exit(1)
		}
		coloured = overlayLogo(coloured, logoImg, *size, *logoFrac)
	}

	if err := savePNG(*output, coloured); err != nil {
		fmt.Fprintf(os.Stderr, "could not save output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("QR code saved to %s\n", *output)
}

// recolour maps black modules to fg and white background to bg.
func recolour(src image.Image, fg, bg color.RGBA, size int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			if lum < 32767 {
				dst.SetRGBA(x, y, fg)
			} else {
				dst.SetRGBA(x, y, bg)
			}
		}
	}
	return dst
}

// overlayLogo centres a scaled logo on the QR inside a circular white pad.
func overlayLogo(base *image.RGBA, logo image.Image, size int, frac float64) *image.RGBA {
	logoSize := int(math.Round(float64(size) * frac))
	if logoSize < 1 {
		return base
	}

	scaled := image.NewRGBA(image.Rect(0, 0, logoSize, logoSize))
	xdraw.BiLinear.Scale(scaled, scaled.Bounds(), logo, logo.Bounds(), xdraw.Over, nil)

	pad    := int(math.Round(float64(logoSize) * 0.12))
	cx, cy := size/2, size/2
	r      := logoSize/2 + pad

	// White circle background
	for y := cy - r - 1; y <= cy+r+1; y++ {
		for x := cx - r - 1; x <= cx+r+1; x++ {
			dx, dy := float64(x-cx), float64(y-cy)
			if dx*dx+dy*dy <= float64(r*r) {
				base.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	// Paste logo
	ox := cx - logoSize/2
	oy := cy - logoSize/2
	draw.Draw(base, image.Rect(ox, oy, ox+logoSize, oy+logoSize), scaled, image.Point{}, draw.Over)

	return base
}

// fetchImage downloads and decodes a PNG or JPEG.
func fetchImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ext := strings.ToLower(filepath.Ext(strings.Split(url, "?")[0]))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Decode(resp.Body)
	default:
		return png.Decode(resp.Body)
	}
}

// savePNG encodes img to a PNG file at path.
func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// parseHex parses "#RRGGBB" → color.RGBA (alpha=255).
func parseHex(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("expected 6 hex digits, got %q", s)
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{r, g, b, 255}, err
}
