package qrgen

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

// sharedHTTPClient is reused across all requests to benefit from connection
// pooling. Per-request timeouts are enforced via context deadlines.
var sharedHTTPClient = &http.Client{}

// qrVersionFromContent determines the QR code version based on content length
// using ISO 18004 byte-mode capacities at error correction level M.
func qrVersionFromContent(content string) int {
	capacities := []int{
		14, 26, 42, 62, 84, 106, 122, 154, 180, 213, // v1-10
		251, 287, 331, 370, 411, 461, 512, 549, 597, 648, // v11-20
		702, 742, 823, 875, 916, 1000, 1062, 1128, 1193, 1267, // v21-30
		1373, 1455, 1541, 1631, 1725, 1812, 1914, 1992, 2102, 2216, // v31-40
	}
	for version, capacity := range capacities {
		if len(content) <= capacity {
			return version + 1
		}
	}
	return 40
}

// isSafeURL rejects loopback, private, and link-local addresses to prevent
// SSRF attacks.
func isSafeURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got %q", u.Scheme)
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil {
		return nil // hostname — allow through
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("logo_url resolves to a disallowed IP range: %s", ip)
	}
	return nil
}

// UrlGet downloads the image at the URL carried in req, saves it to a temp
// file, and returns the temp file path. The caller must remove the file when
// done. maxBytes caps the download to prevent disk/memory exhaustion.
func UrlGet(req *http.Request, maxBytes int64) (string, error) {
	if err := isSafeURL(req.URL.String()); err != nil {
		return "", err
	}

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36")

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return "", fmt.Errorf("URL did not return an image, got Content-Type: %s", ct)
	}

	tmpFile, err := os.CreateTemp("", "logo-*.jpg")
	if err != nil {
		return "", fmt.Errorf("create temp logo file: %w", err)
	}
	defer tmpFile.Close()

	limited := io.LimitReader(resp.Body, maxBytes+1)
	n, err := io.Copy(tmpFile, limited)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("write logo: %w", err)
	}
	if n > maxBytes {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("logo exceeds maximum allowed size of %d bytes", maxBytes)
	}

	return tmpFile.Name(), nil
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

// resizeImage resizes the image at path to size×size and overwrites it as PNG.
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
