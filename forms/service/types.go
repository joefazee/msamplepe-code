package service

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"mime/multipart"
)

// FormDefinitionWithData includes all form data
type FormDefinitionWithData struct {
	FormDefinition       db.FormDefinition     `json:"form_definition"`
	Steps                []db.FormStep         `json:"steps"`
	Fields               []db.FormField        `json:"fields"`
	ExistingData         json.RawMessage       `json:"existing_data,omitempty"`
	SubmissionID         *uuid.UUID            `json:"submission_id"`
	StepProgress         []db.FormStepProgress `json:"step_progress,omitempty"`
	CurrentStep          int32                 `json:"current_step"`
	CompletionPercentage int32                 `json:"completion_percentage"`
}

// FieldOptions for select/radio/checkbox fields
type FieldOptions struct {
	Type    string         `json:"type"`
	Static  []Option       `json:"static,omitempty"`
	Dynamic *DynamicSource `json:"dynamic,omitempty"`
}

// Option for dropdown/radio/checkbox
type Option struct {
	Value string            `json:"value"`
	Label map[string]string `json:"label"`
}

// DynamicSource configuration
type DynamicSource struct {
	SourceID     uuid.UUID         `json:"source_id,omitempty"`
	SourceName   string            `json:"source_name"`
	FilterParams map[string]string `json:"filter_params,omitempty"`
}

// SubmitFormInput for form submission
type SubmitFormInput struct {
	FormID   uuid.UUID                          `json:"form_id"`
	UserID   uuid.UUID                          `json:"user_id"`
	Data     map[string]interface{}             `json:"data"`
	Files    map[string][]*multipart.FileHeader `json:"-"`
	Metadata map[string]interface{}             `json:"metadata"`
}

// FileConfig for file upload validation
type FileConfig struct {
	MaxSize      int64    `json:"max_size"`
	AllowedTypes []string `json:"allowed_types"`
	MaxFiles     int      `json:"max_files,omitempty"`
}

// ValidationRules for field validation
type ValidationRules struct {
	Required    bool                   `json:"required,omitempty"`
	MinLength   *int                   `json:"min_length,omitempty"`
	MaxLength   *int                   `json:"max_length,omitempty"`
	Pattern     *string                `json:"pattern,omitempty"`
	Min         *decimal.Decimal       `json:"min,omitempty"`
	Max         *decimal.Decimal       `json:"max,omitempty"`
	Email       bool                   `json:"email,omitempty"`
	MinItems    *int                   `json:"min_items,omitempty"`
	MaxItems    *int                   `json:"max_items,omitempty"`
	AllRequired bool                   `json:"all_required,omitempty"`
	CustomRules map[string]interface{} `json:"custom_rules,omitempty"`
}

// I18nText for internationalization
type I18nText map[string]string
