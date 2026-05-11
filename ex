package pdfgen

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"

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

var validSizes = map[string]PageSize{
	"A4":     {Width: 8.27, Height: 11.69},
	"Letter": {Width: 8.5, Height: 11},
	"Legal":  {Width: 8.5, Height: 14},
}

var defaultSize = validSizes["A4"]

func parseDimension(value, field string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required when using custom size", field)
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil || f <= 0 {
		return 0, fmt.Errorf("%s '%s' is invalid — must be a positive number, e.g. 6, 8.5", field, value)
	}
	return f, nil
}

func validateSize(size, customWidth, customHeight string) (PageSize, error) {
	usingCustom := customWidth != "" || customHeight != ""

	if usingCustom {
		if size != "" {
			return PageSize{}, fmt.Errorf("size and custom_width/custom_height are mutually exclusive")
		}
		w, err := parseDimension(customWidth, "custom_width")
		if err != nil {
			return PageSize{}, err
		}
		h, err := parseDimension(customHeight, "custom_height")
		if err != nil {
			return PageSize{}, err
		}
		return PageSize{Width: w, Height: h}, nil
	}

	if size == "" {
		return defaultSize, nil
	}

	ps, ok := validSizes[size]
	if !ok {
		var lines []string
		for name, dims := range validSizes {
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

	pageSize, err := validateSize(req.Size, req.CustomWidth, req.CustomHeight)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	slog.Info("pdf request", "filename", filename, "html_length", len(req.HTML), "page_size", pageSize)

	path, err := GenerateNamed(req.HTML, filename, pageSize.Width, pageSize.Height)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Attachment(path, filename)
}
