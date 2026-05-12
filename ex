package chromedp

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
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

var ValidSizes = map[string]PageSize{
	"A4":     {Width: 8.27, Height: 11.69},
	"Letter": {Width: 8.5, Height: 11},
	"Legal":  {Width: 8.5, Height: 14},
}

func ValidateDimension(value, field string) error {
	if value == "" {
		return fmt.Errorf("%s is required when using custom size", field)
	}
	if !strings.HasSuffix(value, "in") {
		return fmt.Errorf("%s '%s' is invalid\nvalid format: e.g. 6, 8.5", field, value)
	}
	numberPart := strings.TrimSuffix(value, "in")
	if _, err := strconv.ParseFloat(numberPart, 64); err != nil {
		return fmt.Errorf("%s '%s' is invalid\nvalid format: e.g. 6, 8.5", field, value)
	}
	return nil
}

func ValidateSize(size string, customWidth, customHeight float64) (PageSize, error) {
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
	ps, ok := ValidSizes[size]
	if !ok {
		var lines []string
		for name, dims := range ValidSizes {
			lines = append(lines, fmt.Sprintf("  %s  width - %.2f  height - %.2f", name, dims.Width, dims.Height))
		}
		sort.Strings(lines)
		return PageSize{}, fmt.Errorf(
			"the size '%s' is invalid\nvalid formats:\n%s\nor provide custom_width and custom_height instead",
			size, strings.Join(lines, "\n"),
		)
	}
	return ps, nil
}

func GenerateHandler(c echo.Context) error {
	reqID := c.Request().Header.Get("X-Request-Id")
	logger := slog.With("request_id", reqID, "remote_ip", c.RealIP())

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error("failed to read request body", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req RequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Error("failed to parse request JSON", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}

	if req.HTML == "" {
		logger.Warn("missing required field", "field", "html")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "html field is required"})
	}
	if req.Filename == "" {
		logger.Warn("missing required field", "field", "filename")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename field is required"})
	}
	if len(req.Filename) > 50 {
		logger.Warn("filename too long", "filename", req.Filename, "length", len(req.Filename))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename must be 50 characters or less"})
	}

	if err := ValidateBody(req.HTML); err != nil {
		logger.Warn("html validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	filename, err := SanitizeFilename(req.Filename)
	if err != nil {
		logger.Warn("filename sanitization failed", "filename", req.Filename, "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	w, h, err := parseFloat(req.CustomWidth, req.CustomHeight)
	if err != nil {
		logger.Warn("invalid custom dimensions", "custom_width", req.CustomWidth, "custom_height", req.CustomHeight, "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid custom dimensions: " + err.Error()})
	}

	pageSize, err := ValidateSize(req.Size, w, h)
	if err != nil {
		logger.Warn("page size validation failed", "size", req.Size, "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	logger.Info("pdf request received", "filename", filename, "html_length", len(req.HTML), "page_size", pageSize)

	path, err := GenerateNamed(req.HTML, filename, pageSize.Width, pageSize.Height)
	if err != nil {
		logger.Error("pdf generation failed", "filename", filename, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	logger.Info("pdf generated successfully", "filename", filename, "path", path)
	return c.Attachment(path, filename)
}

// parseFloat parses custom width/height strings.
// Returns (0, 0, nil) when both are empty so ValidateSize falls back to the named size.
func parseFloat(w, h string) (float64, float64, error) {
	if w == "" && h == "" {
		return 0, 0, nil
	}
	fw, err := strconv.ParseFloat(w, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("custom_width '%s' is not a valid number", w)
	}
	fh, err := strconv.ParseFloat(h, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("custom_height '%s' is not a valid number", h)
	}
	return fw, fh, nil
}
