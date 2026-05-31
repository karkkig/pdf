package qrgen

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type QrRequest struct {
	Content   string `json:"content" example:"https://example.com"`           // Data to encode in QR code
	Dimension int    `json:"dimension" example:"300"`                         // QR image dimension in pixels
	Border    int    `json:"border" example:"4"`                              // Border width around QR code
	LogoURL   string `json:"logo_url" example:"http://localhost:1323/fetchlogo"` // Optional logo image URL
}

// parseQrRequest decodes and validates the QR request from the echo context.
func parseQrRequest(c echo.Context) (*QrRequest, error) {
	var req QrRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("content field is required")
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
// @Summary      Generate QR code with logo
// @Description  Generates a QR code using the yeqown library with optional logo embedding
// @Tags         QRCode
// @Accept       json
// @Produce      image/png
// @Param        request  body      QrRequest  true  "QR generation request payload"
// @Success      200      {file}    file        "Generated QR code image"
// @Failure      400      {object}  map[string]string "Bad request"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Router       /qr1 [post]
func Qr1Handler(c echo.Context) error {
	req, err := parseQrRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	outPath, err := CreateQRWithLogo(req.Content, req.LogoURL, req.Dimension, req.Border)
	if err != nil {
		slog.Error("failed to create QR code with logo", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create QR code: " + err.Error()})
	}
	return c.Attachment(outPath, "qrcode_with_logo.png")
}

// Qr2Handler
// @Summary      Generate QR code with logo
// @Description  Generates a QR code using the qr-code library with optional logo embedding
// @Tags         QRCode
// @Accept       json
// @Produce      image/png
// @Param        request  body      QrRequest  true  "QR generation request payload"
// @Success      200      {file}    file        "Generated QR code image"
// @Failure      400      {object}  map[string]string "Bad request"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Router       /qr2 [post]
func Qr2Handler(c echo.Context) error {
	req, err := parseQrRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	outPath, err := CreateQRWithLogo2(req.Content, req.LogoURL, req.Dimension, req.Border)
	if err != nil {
		slog.Error("failed to create QR code with logo", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create QR code: " + err.Error()})
	}
	return c.Attachment(outPath, "qrcode_with_logo.png")
}

// FetchLogoHandler
// @Summary      Fetch sample logo image
// @Description  Returns a sample logo image for testing QR code generation
// @Tags         QRCode
// @Produce      image/jpeg
// @Success      200  {file}    file "Sample logo image"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /fetchlogo [get]
func FetchLogoHandler(c echo.Context) error {
	return c.Attachment("logo.jpg", "download.jpg")
}
