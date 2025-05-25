package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/timchuks/monieverse/internal/config"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/validator"
)

// FileValidationMode determines how strict file validation should be
type FileValidationMode string

const (
	FileValidationModeDraft   FileValidationMode = "draft"   // Lenient validation for drafts
	FileValidationModePartial FileValidationMode = "partial" // Basic validation for partial updates
	FileValidationModeStep    FileValidationMode = "step"    // Step-specific validation
	FileValidationModeFinal   FileValidationMode = "final"   // Strict validation for final submission
)

// FileValidationContext provides context for file validation
type FileValidationContext struct {
	Mode           FileValidationMode
	StepNumber     *int32
	AllowedFields  map[string]bool // Which file fields are allowed in this context
	RequiredFields map[string]bool // Which file fields are required in this context
}

// FileValidationConfig represents file validation rules from form field config
type FileValidationConfig struct {
	MaxSize           int64    `json:"max_size"`           // Maximum file size in bytes
	AllowedTypes      []string `json:"allowed_types"`      // Allowed MIME types
	MaxFiles          int      `json:"max_files"`          // Maximum number of files
	RequiredInDraft   bool     `json:"required_in_draft"`  // Whether required even in draft mode
	AllowedExtensions []string `json:"allowed_extensions"` // Allowed file extensions
	MinFiles          int      `json:"min_files"`          // Minimum number of files
	ScanForVirus      bool     `json:"scan_for_virus"`     // Whether to scan for viruses
	ValidateContent   bool     `json:"validate_content"`   // Whether to validate file content
}

// FileValidationResult contains validation results for a file
type FileValidationResult struct {
	Valid    bool               `json:"valid"`
	Errors   []string           `json:"errors"`
	Warnings []string           `json:"warnings"`
	FileInfo *ValidatedFileInfo `json:"file_info,omitempty"`
}

// ValidatedFileInfo contains information about a validated file
type ValidatedFileInfo struct {
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	Extension    string `json:"extension"`
	SHA256Hash   string `json:"sha256_hash"`
	IsClean      bool   `json:"is_clean"` // Virus scan result
}

// FileValidator handles comprehensive file validation
type FileValidator struct {
	config           *config.Config
	virusScanner     VirusScanner     // Interface for virus scanning
	contentValidator ContentValidator // Interface for content validation
}

// VirusScanner interface for virus scanning implementations
type VirusScanner interface {
	ScanFile(file multipart.File) (bool, error)
}

// ContentValidator interface for file content validation
type ContentValidator interface {
	ValidateContent(file multipart.File, expectedType string) (bool, error)
}

// NewFileValidator creates a new file validator
func NewFileValidator(config *config.Config, virusScanner VirusScanner, contentValidator ContentValidator) *FileValidator {
	return &FileValidator{
		config:           config,
		virusScanner:     virusScanner,
		contentValidator: contentValidator,
	}
}

// ValidateFiles validates all uploaded files according to form field configurations
func (s *FormService) ValidateFiles(
	fields []db.FormField,
	files map[string][]*multipart.FileHeader,
	ctx FileValidationContext,
) map[string]FileValidationResult {

	results := make(map[string]FileValidationResult)

	// Validate each file field
	for _, field := range fields {
		if field.FieldType != "file" && field.FieldType != "files" {
			continue
		}

		fieldName := field.FieldName
		fileHeaders := files[fieldName]

		// Parse file configuration
		var fileConfig FileValidationConfig
		if field.FileConfig != nil {
			if err := json.Unmarshal(field.FileConfig, &fileConfig); err != nil {
				results[fieldName] = FileValidationResult{
					Valid:  false,
					Errors: []string{fmt.Sprintf("Invalid file configuration: %v", err)},
				}
				continue
			}
		}

		// Apply default configuration if not specified
		s.applyDefaultFileConfig(&fileConfig)

		// Validate field-level requirements
		result := s.validateFileField(field, fileHeaders, fileConfig, ctx)
		results[fieldName] = result
	}

	return results
}

// validateFileField validates a specific file field
func (s *FormService) validateFileField(
	field db.FormField,
	fileHeaders []*multipart.FileHeader,
	config FileValidationConfig,
	ctx FileValidationContext,
) FileValidationResult {

	result := FileValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	fieldName := field.FieldName

	// Check if field is required in current context
	isRequired := s.isFileFieldRequired(field, config, ctx)

	// Check if files are provided
	if len(fileHeaders) == 0 {
		if isRequired {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("%s is required", fieldName))
		}
		return result
	}

	// Validate file count
	if err := s.validateFileCount(fileHeaders, config, ctx.Mode); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	// Validate each individual file
	for i, fileHeader := range fileHeaders {
		fileResult := s.validateIndividualFile(fileHeader, config, ctx)

		if !fileResult.Valid {
			result.Valid = false
			for _, err := range fileResult.Errors {
				result.Errors = append(result.Errors, fmt.Sprintf("File %d (%s): %s", i+1, fileHeader.Filename, err))
			}
		}

		// Add warnings
		for _, warning := range fileResult.Warnings {
			result.Warnings = append(result.Warnings, fmt.Sprintf("File %d (%s): %s", i+1, fileHeader.Filename, warning))
		}
	}

	return result
}

// validateIndividualFile validates a single file
func (s *FormService) validateIndividualFile(
	fileHeader *multipart.FileHeader,
	config FileValidationConfig,
	ctx FileValidationContext,
) FileValidationResult {

	result := FileValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Basic file info validation
	if fileHeader.Size == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "empty file not allowed")
		return result
	}

	// Size validation
	if config.MaxSize > 0 && fileHeader.Size > config.MaxSize {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file size %d bytes exceeds maximum %d bytes", fileHeader.Size, config.MaxSize))
		return result
	}

	// Extension validation
	if len(config.AllowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !s.contains(config.AllowedExtensions, ext) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("file extension %s not allowed", ext))
			return result
		}
	}

	// Open file for content-based validation
	file, err := fileHeader.Open()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("cannot open file: %v", err))
		return result
	}
	defer file.Close()

	// MIME type validation (read from file content, not just header)
	mimeType, err := s.detectMimeType(file)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("cannot detect file type: %v", err))
		return result
	}

	if len(config.AllowedTypes) > 0 && !s.contains(config.AllowedTypes, mimeType) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file type %s not allowed", mimeType))
		return result
	}

	// Create file info
	fileInfo := &ValidatedFileInfo{
		OriginalName: fileHeader.Filename,
		Size:         fileHeader.Size,
		MimeType:     mimeType,
		Extension:    filepath.Ext(fileHeader.Filename),
	}

	// Calculate file hash for integrity
	if hash, err := s.calculateFileHash(file); err == nil {
		fileInfo.SHA256Hash = hash
	} else {
		result.Warnings = append(result.Warnings, "could not calculate file hash")
	}

	// Virus scanning (only for final submissions or if specifically required)
	if (ctx.Mode == FileValidationModeFinal || config.ScanForVirus) && s.fileValidator != nil && s.fileValidator.virusScanner != nil {
		// Reset file pointer
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		isClean, err := s.fileValidator.virusScanner.ScanFile(file)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("virus scan failed: %v", err))
		} else if !isClean {
			result.Valid = false
			result.Errors = append(result.Errors, "file failed virus scan")
			return result
		}
		fileInfo.IsClean = isClean
	}

	// Content validation (only for final submissions or if specifically required)
	if (ctx.Mode == FileValidationModeFinal || config.ValidateContent) && s.fileValidator != nil && s.fileValidator.contentValidator != nil {
		// Reset file pointer
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		isValid, err := s.fileValidator.virusScanner.ScanFile(file)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("content validation failed: %v", err))
		} else if !isValid {
			result.Valid = false
			result.Errors = append(result.Errors, "file content validation failed")
			return result
		}
	}

	result.FileInfo = fileInfo
	return result
}

// isFileFieldRequired determines if a file field is required in the current context
func (s *FormService) isFileFieldRequired(field db.FormField, config FileValidationConfig, ctx FileValidationContext) bool {
	switch ctx.Mode {
	case FileValidationModeDraft:
		// In draft mode, only require if explicitly marked as required in draft
		return config.RequiredInDraft

	case FileValidationModePartial:
		// In partial mode, not required unless specifically marked
		return config.RequiredInDraft

	case FileValidationModeStep:
		// In step mode, check if this field is in the current step and required
		if ctx.RequiredFields != nil {
			return ctx.RequiredFields[field.FieldName]
		}
		return field.IsRequired

	case FileValidationModeFinal:
		// In final mode, use the field's required setting
		return field.IsRequired

	default:
		return field.IsRequired
	}
}

// validateFileCount validates the number of files uploaded
func (s *FormService) validateFileCount(fileHeaders []*multipart.FileHeader, config FileValidationConfig, mode FileValidationMode) error {
	fileCount := len(fileHeaders)

	// Check minimum files (usually only for final submissions)
	if mode == FileValidationModeFinal && config.MinFiles > 0 && fileCount < config.MinFiles {
		return fmt.Errorf("minimum %d files required, got %d", config.MinFiles, fileCount)
	}

	// Check maximum files (applies to all modes)
	if config.MaxFiles > 0 && fileCount > config.MaxFiles {
		return fmt.Errorf("maximum %d files allowed, got %d", config.MaxFiles, fileCount)
	}

	return nil
}

// detectMimeType detects MIME type from file content (more secure than trusting headers)
func (s *FormService) detectMimeType(file multipart.File) (string, error) {
	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Reset file pointer
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	return http.DetectContentType(buffer[:n]), nil
}

// calculateFileHash calculates SHA256 hash of file content
func (s *FormService) calculateFileHash(file multipart.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	// Reset file pointer
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// applyDefaultFileConfig applies default configuration values
func (s *FormService) applyDefaultFileConfig(config *FileValidationConfig) {
	if config.MaxSize == 0 {
		config.MaxSize = s.config.MaxFileUploadSize // Use global default
	}

	if len(config.AllowedTypes) == 0 {
		config.AllowedTypes = s.config.AllowedMimeTypes // Use global defaults
	}

	if config.MaxFiles == 0 {
		config.MaxFiles = 10 // Reasonable default
	}
}

// ValidateSubmissionWithFiles validates both regular form fields and associated file uploads using validation contexts.
func (s *FormService) ValidateSubmissionWithFiles(
	fields []db.FormField,
	data map[string]interface{},
	files map[string][]*multipart.FileHeader,
	ctx ValidationContext,
) error {

	// First validate regular fields
	if err := s.ValidateSubmission(fields, data, ctx); err != nil {
		return err
	}

	// Then validate files
	fileCtx := FileValidationContext{
		Mode:       FileValidationMode(ctx.Mode), // Convert validation mode
		StepNumber: ctx.StepNumber,
	}

	fileResults := s.ValidateFiles(fields, files, fileCtx)

	// Check for file validation errors
	v := validator.New()
	for fieldName, result := range fileResults {
		if !result.Valid {
			for _, err := range result.Errors {
				v.AddError(fieldName, err)
			}
		}

		// Log warnings but don't fail validation
		if len(result.Warnings) > 0 {
			s.logger.Debug("File validation warnings", map[string]interface{}{
				"field":    fieldName,
				"warnings": result.Warnings,
			})
		}
	}

	if !v.Valid() {
		return validator.NewValidationError("file validation failed", v.Errors)
	}

	return nil
}
