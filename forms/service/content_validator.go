package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
)

// PDFContentValidator validates PDF file structure
type PDFContentValidator struct{}

// NewPDFContentValidator creates a new PDF content validator
func NewPDFContentValidator() *PDFContentValidator {
	return &PDFContentValidator{}
}

// ValidateContent validates PDF file content
func (p *PDFContentValidator) ValidateContent(file multipart.File, expectedType string) (bool, error) {
	if !strings.HasPrefix(expectedType, "application/pdf") {
		return true, nil
	}

	// Read PDF header
	header := make([]byte, 8)
	n, err := file.Read(header)
	if err != nil {
		return false, fmt.Errorf("failed to read file header: %w", err)
	}

	// Reset file pointer
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	if n < 4 || !bytes.Equal(header[:4], []byte("%PDF")) {
		return false, nil // Not a valid PDF
	}

	return true, nil
}

// ImageContentValidator validates image file structure
type ImageContentValidator struct{}

// NewImageContentValidator creates a new image content validator
func NewImageContentValidator() *ImageContentValidator {
	return &ImageContentValidator{}
}

// ValidateContent validates image file content
func (i *ImageContentValidator) ValidateContent(file multipart.File, expectedType string) (bool, error) {
	if !strings.HasPrefix(expectedType, "image/") {
		return true, nil // Not an image, skip validation
	}

	header := make([]byte, 12)
	n, err := file.Read(header)
	if err != nil {
		return false, fmt.Errorf("failed to read file header: %w", err)
	}

	// Reset file pointer
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	if n < 4 {
		return false, nil
	}

	switch expectedType {
	case "image/jpeg":
		return bytes.Equal(header[:2], []byte{0xFF, 0xD8}), nil
	case "image/png":
		return bytes.Equal(header[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}), nil
	case "image/gif":
		return bytes.Equal(header[:6], []byte("GIF87a")) || bytes.Equal(header[:6], []byte("GIF89a")), nil
	default:
		return true, nil
	}
}

// CompositeContentValidator combines multiple validators
type CompositeContentValidator struct {
	validators map[string]ContentValidator
	fallback   ContentValidator
}

// NewCompositeContentValidator creates a composite validator
func NewCompositeContentValidator() *CompositeContentValidator {
	return &CompositeContentValidator{
		validators: make(map[string]ContentValidator),
		fallback:   &NoOpContentValidator{},
	}
}

// AddValidator adds a validator for specific MIME type pattern
func (c *CompositeContentValidator) AddValidator(mimePattern string, validator ContentValidator) {
	c.validators[mimePattern] = validator
}

// SetFallback sets the fallback validator
func (c *CompositeContentValidator) SetFallback(validator ContentValidator) {
	c.fallback = validator
}

// ValidateContent validates using appropriate validator
func (c *CompositeContentValidator) ValidateContent(file multipart.File, expectedType string) (bool, error) {
	// Find matching validator
	for pattern, validator := range c.validators {
		if strings.Contains(expectedType, pattern) {
			return validator.ValidateContent(file, expectedType)
		}
	}

	// Use fallback validator
	return c.fallback.ValidateContent(file, expectedType)
}

// NoOpContentValidator does no validation (always passes)
type NoOpContentValidator struct{}

// ValidateContent always returns true
func (n *NoOpContentValidator) ValidateContent(file multipart.File, expectedType string) (bool, error) {
	return true, nil
}

// CreateContentValidator creates appropriate content validator
func CreateContentValidator() ContentValidator {
	composite := NewCompositeContentValidator()

	// Add specific validators
	composite.AddValidator("application/pdf", NewPDFContentValidator())
	composite.AddValidator("image/", NewImageContentValidator())

	composite.SetFallback(&NoOpContentValidator{})

	return composite
}
