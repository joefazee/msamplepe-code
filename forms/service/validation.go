package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"regexp"

	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/validator"
)

// ValidationMode defines how strict validation should be
type ValidationMode string

const (
	ValidationModeFull    ValidationMode = "full"    // All required fields must be present
	ValidationModePartial ValidationMode = "partial" // Only validate provided fields
	ValidationModeStep    ValidationMode = "step"    // Validate specific step only
)

// ValidationContext provides context for validation
type ValidationContext struct {
	Mode           ValidationMode
	StepNumber     *int32                 // For step validation
	ProvidedFields map[string]interface{} // Fields being validated
}

// ValidateSubmission validates form data based on context
func (s *FormService) ValidateSubmission(fields []db.FormField, data map[string]interface{}, ctx ValidationContext) error {
	v := validator.New()

	switch ctx.Mode {
	case ValidationModeFull:
		return s.validateFull(v, fields, data)
	case ValidationModePartial:
		return s.validatePartial(v, fields, data)
	case ValidationModeStep:
		return s.validateStep(v, fields, data, *ctx.StepNumber)
	default:
		return s.validateFull(v, fields, data)
	}
}

// validateFull performs complete validation (original logic)
func (s *FormService) validateFull(v *validator.Validator, fields []db.FormField, data map[string]interface{}) error {
	for _, field := range fields {
		value, exists := data[field.FieldName]

		// Check required fields
		if field.IsRequired && (!exists || s.isEmpty(value)) {
			v.AddError(field.FieldName, "field is required")
			continue
		}

		if !exists || value == nil {
			continue
		}

		// Validate field content
		if err := s.validateFieldValue(v, field, value); err != nil {
			return err
		}
	}

	if !v.Valid() {
		return validator.NewValidationError("validation failed", v.Errors)
	}
	return nil
}

// validatePartial only validates provided fields (for drafts)
func (s *FormService) validatePartial(v *validator.Validator, fields []db.FormField, data map[string]interface{}) error {
	for _, field := range fields {
		value, exists := data[field.FieldName]

		// Skip missing fields in partial validation
		if !exists || value == nil {
			continue
		}

		// Validate only provided fields
		if err := s.validateFieldValue(v, field, value); err != nil {
			return err
		}
	}

	if !v.Valid() {
		return validator.NewValidationError("validation failed", v.Errors)
	}
	return nil
}

// validateStep validates fields for a specific step
func (s *FormService) validateStep(v *validator.Validator, fields []db.FormField, data map[string]interface{}, stepNumber int32) error {
	stepFields := make([]db.FormField, 0)
	for _, field := range fields {
		if field.FormStepID != uuid.Nil {
			stepFields = append(stepFields, field)
		}
	}

	for _, field := range stepFields {
		value, exists := data[field.FieldName]

		if field.IsRequired && (!exists || s.isEmpty(value)) {
			v.AddError(field.FieldName, "field is required for this step")
			continue
		}

		if exists && value != nil {
			if err := s.validateFieldValue(v, field, value); err != nil {
				return err
			}
		}
	}

	if !v.Valid() {
		return validator.NewValidationError("step validation failed", v.Errors)
	}
	return nil
}

// validateFieldValue validates individual field values
func (s *FormService) validateFieldValue(v *validator.Validator, field db.FormField, value interface{}) error {
	// Parse validation rules
	var rules ValidationRules
	if field.ValidationRules != nil {
		if err := json.Unmarshal(field.ValidationRules, &rules); err != nil {
			return fmt.Errorf("failed to parse validation rules for field %s: %w", field.FieldName, err)
		}
	}

	// Type-specific validation
	switch field.FieldType {
	case "email":
		if str, ok := value.(string); ok {
			v.Check(validator.IsEmail(str), field.FieldName, "invalid email format")
		}

	case "text", "textarea":
		if str, ok := value.(string); ok {
			if rules.MinLength != nil {
				v.Check(len(str) >= *rules.MinLength, field.FieldName,
					fmt.Sprintf("must be at least %d characters", *rules.MinLength))
			}
			if rules.MaxLength != nil {
				v.Check(len(str) <= *rules.MaxLength, field.FieldName,
					fmt.Sprintf("must be at most %d characters", *rules.MaxLength))
			}
			if rules.Pattern != nil {
				matched, _ := regexp.MatchString(*rules.Pattern, str)
				v.Check(matched, field.FieldName, "invalid format")
			}
		}

	case "number", "currency":
		if num, ok := s.toFloat64(value); ok {
			if rules.Min != nil {
				minVal, _ := rules.Min.Float64()
				v.Check(num >= minVal, field.FieldName,
					fmt.Sprintf("must be at least %v", minVal))
			}
			if rules.Max != nil {
				maxVal, _ := rules.Max.Float64()
				v.Check(num <= maxVal, field.FieldName,
					fmt.Sprintf("must be at most %v", maxVal))
			}
		}

	case "select", "radio":
		var options FieldOptions
		if field.Options != nil {
			_ = json.Unmarshal(field.Options, &options)
		}

		if options.Type == "static" && len(options.Static) > 0 {
			validOptions := make([]string, len(options.Static))
			for i, opt := range options.Static {
				validOptions[i] = opt.Value
			}
			if str, ok := value.(string); ok {
				v.Check(validator.In(str, validOptions...), field.FieldName, "invalid option")
			}
		}

	case "file", "files":
		// File validation would be handled separately during upload
		// Here we might just check if file references are valid
		break
	}

	return nil
}
