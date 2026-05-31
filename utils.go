package qrgen

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

// qrVersionFromContent determines the QR code version based on the content length
// using ISO 18004 byte-mode capacities at error correction level M.
func qrVersionFromContent(content string) int {
	contentLen := len(content)

	capacities := []int{
		14, 26, 42, 62, 84, 106, 122, 154, 180, 213, // v1-10
		251, 287, 331, 370, 411, 461, 512, 549, 597, 648, // v11-20
		702, 742, 823, 875, 916, 1000, 1062, 1128, 1193, 1267, // v21-30
		1373, 1455, 1541, 1631, 1725, 1812, 1914, 1992, 2102, 2216, // v31-40
	}

	for version, capacity := range capacities {
		if contentLen <= capacity {
			return version + 1 // versions are 1-indexed
		}
	}
	return 40 // max version
}

// resizeLogoToTarget resizes the logo at path to targetSize×targetSize in place.
func resizeLogoToTarget(path string, targetSize int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("decode logo: %w", err)
	}

	resized := resize.Resize(uint(targetSize), uint(targetSize), img, resize.Lanczos3)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, resized, &jpeg.Options{Quality: 95})
}

// resizeImage resizes the image at path to size×size and overwrites the file as PNG.
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

// UrlGet downloads the image at url, saves it to a temp file, and returns the
// temp file path. The caller is responsible for removing the file when done.
func UrlGet(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Reject non-image responses early to avoid writing garbage to disk.
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return "", fmt.Errorf("URL did not return an image, got Content-Type: %s", ct)
	}

	tmpFile, err := os.CreateTemp("", "logo-*.jpg")
	if err != nil {
		return "", fmt.Errorf("create temp logo file: %w", err)
	}
	defer tmpFile.Close()

	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("write logo: %w", err)
	}

	return tmpFile.Name(), nil
}
