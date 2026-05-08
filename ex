package chromedp

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

type RequestBody struct {
	Template string         `json:"template"`
	Data     map[string]any `json:"data"`
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

	if req.Template == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "template field is required"})
	}

	if err := ValidateBody(req.Template); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	tmpl, err := template.New("doc").Parse(req.Template)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "template parse error: " + err.Error()})
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, req.Data); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "template render error: " + err.Error()})
	}

	rawFilename := c.Param("filename")
	filename, err := SanitizeFilename(rawFilename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	slog.Info("pdf request", "filename", filename, "html_length", rendered.Len())

	path, err := GenerateNamed(rendered.String(), filename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer os.Remove(path)

	return c.Attachment(path, filename)
}
