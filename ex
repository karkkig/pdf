package pdfgen

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
	Width  string
	Height string
}

var ValidSizes = map[string]PageSize{
	"A4":     {Width: "8.27in", Height: "11.69in"},
	"Letter": {Width: "8.5in", Height: "11in"},
	"Legal":  {Width: "8.5in", Height: "14in"},
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

func normalizeToInches(value string) string {
	return value + "in"
}

func ValidateSize(size, customWidth, customHeight string) (PageSize, error) {
	if customWidth != "" || customHeight != "" {
		if size != "" {
			return PageSize{}, fmt.Errorf("size and custom_width/custom_height are mutually exclusive")
		}
		if err := ValidateDimension(customWidth, "custom_width"); err != nil {
			return PageSize{}, err
		}
		if err := ValidateDimension(customHeight, "custom_height"); err != nil {
			return PageSize{}, err
		}
		return PageSize{
			Width:  normalizeToInches(customWidth),
			Height: normalizeToInches(customHeight),
		}, nil
	}

	if size == "" {
		size = "A4"
	}
	ps, ok := ValidSizes[size]
	if !ok {
		var lines []string
		for name, dims := range ValidSizes {
			lines = append(lines, fmt.Sprintf("  %s  width - %s  height - %s", name, dims.Width, dims.Height))
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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req RequestBody
	if err := json.Unmarshal(body, &req); err != nil {
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

	pageSize, err := ValidateSize(req.Size, req.CustomWidth, req.CustomHeight)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	slog.Info("pdf request", "filename", filename, "html_length", len(req.HTML), "page_size", pageSize)

	path, err := GenerateNamed(req.HTML, filename, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Attachment(path, filename)
}
