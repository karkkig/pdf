package pdf

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrMissingFilename = errors.New("X-PDF-Name header is required")
	ErrEmptyBody       = errors.New("HTML body must not be empty")
)

// SanitizeFilename cleans the provided name, ensures it has a .pdf
// extension, and strips any path components to prevent traversal.
func SanitizeFilename(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", ErrMissingFilename
	}

	name = filepath.Base(name) // strip any directory component
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		name += ".pdf"
	}

	return name, nil
}

// ValidateBody returns ErrEmptyBody when the HTML string is blank.
func ValidateBody(html string) error {
	if strings.TrimSpace(html) == "" {
		return ErrEmptyBody
	}
	return nil
}
