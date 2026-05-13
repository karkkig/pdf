package chromedp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	errMissingFilename = errors.New("filename field is required")
	errEmptyHTML       = errors.New("html field is required")
)

type RequestBody struct {
	HTML         string `json:"html"`
	Filename     string `json:"filename"`
	Size         string `json:"size"`
	CustomWidth  string `json:"custom_width"`
	CustomHeight string `json:"custom_height"`
}

type PageSize struct {
	Width  float64
	Height float64
}

var validSizes = map[string]PageSize{
	"A4":     {Width: 8.27, Height: 11.69},
	"Letter": {Width: 8.5, Height: 11},
	"Legal":  {Width: 8.5, Height: 14},
}

func sanitizeFilename(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errMissingFilename
	}
	name = filepath.Base(name)
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		name += ".pdf"
	}
	return name, nil
}

func validateSize(size string, customWidth, customHeight float64) (PageSize, error) {
	if customWidth != 0 || customHeight != 0 {
		if size != "" {
			return PageSize{}, fmt.Errorf("size and custom_width/custom_height are mutually exclusive")
		}
		if customWidth <= 0 || customHeight <= 0 {
			return PageSize{}, fmt.Errorf("custom_width and custom_height must be positive")
		}
		return PageSize{Width: customWidth, Height: customHeight}, nil
	}

	if size == "" {
		size = "A4"
	}
	ps, ok := validSizes[size]
	if !ok {
		var lines []string
		for name, dims := range validSizes {
			lines = append(lines, fmt.Sprintf("  %s  width - %.2f  height - %.2f", name, dims.Width, dims.Height))
		}
		sort.Strings(lines)
		return PageSize{}, fmt.Errorf(
			"invalid size '%s'\nvalid sizes:\n%s\nor provide custom_width and custom_height instead",
			size, strings.Join(lines, "\n"),
		)
	}
	return ps, nil
}

func parseCustomDimensions(w, h string) (float64, float64, error) {
	if w == "" && h == "" {
		return 0, 0, nil
	}
	if w == "" || h == "" {
		return 0, 0, fmt.Errorf("custom_width and custom_height must both be provided together")
	}
	fw, err := strconv.ParseFloat(w, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid custom_width '%s': must be a number (e.g. 8.5)", w)
	}
	fh, err := strconv.ParseFloat(h, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid custom_height '%s': must be a number (e.g. 11)", h)
	}
	return fw, fh, nil
}

func GenerateHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req RequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("failed to parse request JSON", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}

	if strings.TrimSpace(req.HTML) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errEmptyHTML.Error()})
	}
	if len(req.Filename) > 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename must be 50 characters or less"})
	}

	filename, err := sanitizeFilename(req.Filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	w, h, err := parseCustomDimensions(req.CustomWidth, req.CustomHeight)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	pageSize, err := validateSize(req.Size, w, h)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	slog.Info("pdf request", "filename", filename, "html_length", len(req.HTML), "page_size", pageSize)

	path, err := GenerateNamed(req.HTML, filename, pageSize.Width, pageSize.Height)
	if err != nil {
		slog.Error("pdf generation failed", "filename", filename, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Attachment(path, filename)
}

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"time":   time.Now().Format("2006-01-02 15:04:05 Monday"),
		"status": "OK",
	})
}
