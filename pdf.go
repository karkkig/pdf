package handler

import (
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"your_module/internal/pdf"
)

// GeneratePDF handles POST /generate-pdf.
//
// Headers:
//
//	X-PDF-Name: invoice.pdf   (required; .pdf appended if missing)
//
// Body:
//
//	Raw HTML to render into the PDF.
func GeneratePDF(c echo.Context) error {
	filename, htmlContent, err := parseRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	outputPath, err := pdf.GenerateNamed(htmlContent, filename)
	if err != nil {
		c.Logger().Errorf("pdf generation failed: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to generate PDF: " + err.Error(),
		})
	}
	defer os.Remove(outputPath)

	return c.Attachment(outputPath, filename)
}

// parseRequest reads and validates the incoming header and body.
func parseRequest(c echo.Context) (filename, html string, err error) {
	filename, err = pdf.SanitizeFilename(c.Request().Header.Get("X-PDF-Name"))
	if err != nil {
		return
	}

	raw, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return "", "", errors.New("could not read request body")
	}
	html = string(raw)

	err = pdf.ValidateBody(html)
	return
}
