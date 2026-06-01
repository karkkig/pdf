package qrgen

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Config holds tunable parameters for the QR generation service.
type Config struct {
	HTTPTimeout     time.Duration
	MaxLogoBytes    int64
	LogoSizeRatio   float64
	MaxContentBytes int
}

// DefaultConfig returns a Config with sensible production defaults.
func DefaultConfig() Config {
	return Config{
		HTTPTimeout:     10 * time.Second,
		MaxLogoBytes:    5 << 20, // 5 MB
		LogoSizeRatio:   0.20,
		MaxContentBytes: 2953, // QR v40, byte mode, level M
	}
}

// QrRequest is the JSON body accepted by every QR generation endpoint.
type QrRequest struct {
	Content   string `json:"content"   example:"https://example.com"`             // Data to encode
	Dimension int    `json:"dimension" example:"300"`                              // Output image size in pixels
	Border    int    `json:"border"    example:"4"`                                // Quiet-zone width in modules
	LogoURL   string `json:"logo_url"  example:"http://localhost:1323/fetchlogo"`  // Optional logo URL
}

// parseQrRequest decodes and validates the QR request body.
func parseQrRequest(c echo.Context, maxContent int) (*QrRequest, error) {
	var req QrRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("content field is required")
	}
	if len(req.Content) > maxContent {
		return nil, fmt.Errorf("content too long: max %d bytes", maxContent)
	}
	if req.Dimension <= 50 {
		return nil, fmt.Errorf("dimension must be greater than 50")
	}
	if req.Border < 0 {
		return nil, fmt.Errorf("border must be non-negative")
	}
	return &req, nil
}

// Qr1Handler
// @Summary      Generate QR code with logo (yeqown)
// @Description  Generates a QR code using the yeqown library with optional logo embedding
// @Tags         QRCode
// @Accept       json
// @Produce      image/png
// @Param        request  body      QrRequest  true  "QR generation request payload"
// @Success      200      {file}    file                    "Generated QR code image"
// @Failure      400      {object}  map[string]string       "Bad request"
// @Failure      500      {object}  map[string]string       "Internal server error"
// @Router       /qr1 [post]
func Qr1Handler(cfg Config) func(c echo.Context) error {
	return func(c echo.Context) error {
		req, err := parseQrRequest(c, cfg.MaxContentBytes)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		outPath, err := CreateQRWithLogo(c.Request().Context(), req.Content, req.LogoURL, req.Dimension, req.Border, cfg)
		if err != nil {
			slog.Error("failed to create QR code (yeqown)", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create QR code: " + err.Error()})
		}
		defer os.Remove(outPath)
		return c.Attachment(outPath, "qrcode_with_logo.png")
	}
}

// Qr2Handler
// @Summary      Generate QR code with logo (qr-gode)
// @Description  Generates a QR code using the qr-gode library with optional logo embedding
// @Tags         QRCode
// @Accept       json
// @Produce      image/png
// @Param        request  body      QrRequest  true  "QR generation request payload"
// @Success      200      {file}    file                    "Generated QR code image"
// @Failure      400      {object}  map[string]string       "Bad request"
// @Failure      500      {object}  map[string]string       "Internal server error"
// @Router       /qr2 [post]
func Qr2Handler(cfg Config) func(c echo.Context) error {
	return func(c echo.Context) error {
		req, err := parseQrRequest(c, cfg.MaxContentBytes)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		outPath, err := CreateQRWithLogo2(c.Request().Context(), req.Content, req.LogoURL, req.Dimension, req.Border, cfg)
		if err != nil {
			slog.Error("failed to create QR code (qr-gode)", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create QR code: " + err.Error()})
		}
		defer os.Remove(outPath)
		return c.Attachment(outPath, "qrcode_with_logo.png")
	}
}

// FetchLogoHandler returns a sample logo image for testing.
//
// @Summary      Fetch sample logo image
// @Description  Returns a sample logo image for testing QR code generation
// @Tags         QRCode
// @Produce      image/jpeg
// @Success      200  {file}    file              "Sample logo image"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /fetchlogo [get]
func FetchLogoHandler(c echo.Context) error {
	return c.Attachment("logo.jpg", "download.jpg")
}

// HealthHandler reports service liveness.
//
// @Summary      Health check
// @Description  Returns 200 OK when the service is up
// @Tags         Ops
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
