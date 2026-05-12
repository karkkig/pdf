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

	if req.HTML == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "html field is required"})
	}
	if req.Filename == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename field is required"})
	}
	if len(req.Filename) > 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "filename must be 50 characters or less"})
	}

	if err := ValidateBody(req.HTML); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	filename, err := SanitizeFilename(req.Filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	w, h := parseFloat(req.CustomWidth, req.CustomHeight)
	pageSize, err := ValidateSize(req.Size, w, h)
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

func parseFloat(w, h string) (float64, float64) {
	fw, err := strconv.ParseFloat(w, 64)
	fh, err := strconv.ParseFloat(h, 64)
	if err != nil {
		return 8.27, 11.69 // A4 size default
	}
	return fw, fh
}
