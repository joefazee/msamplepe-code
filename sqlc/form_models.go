package db

import (
	"encoding/json"
	"github.com/google/uuid"
)

// FormDefinitionInput for creating forms
type FormDefinitionInput struct {
	Name                      string                  `json:"name"`
	Slug                      string                  `json:"slug"`
	Description               string                  `json:"description"`
	FormType                  string                  `json:"form_type"`
	IsMultiStep               bool                    `json:"is_multi_step"`
	RequiresApproval          bool                    `json:"requires_approval"`
	IsEditableAfterSubmission bool                    `json:"is_editable_after_submission"`
	CreatedBy                 uuid.UUID               `json:"created_by"`
	ApprovalWorkflow          *ApprovalWorkflowInput  `json:"approval_workflow"`
	Steps                     []StepInput             `json:"steps"`
	Fields                    []FieldInput            `json:"fields"`
	PersistenceConfig         *PersistenceConfigInput `json:"persistence_config"`
}

// FormSubmissionUpdateInput for updating submissions
type FormSubmissionUpdateInput struct {
	SubmissionID     uuid.UUID                 `json:"submission_id"`
	FormDefinitionID uuid.UUID                 `json:"form_definition_id"`
	UserID           uuid.UUID                 `json:"user_id"`
	Status           string                    `json:"status"`
	Data             map[string]interface{}    `json:"data"`
	Files            []FormSubmissionFileInput `json:"files"`
	Metadata         map[string]interface{}    `json:"metadata"`
	IsPartialUpdate  bool                      `json:"is_partial_update"`
	ExistingData     json.RawMessage           `json:"existing_data"`
}

type ApprovalWorkflowInput struct {
	States      []map[string]interface{} `json:"states"`
	Transitions []map[string]interface{} `json:"transitions"`
}

type StepInput struct {
	StepNumber  int    `json:"step_number"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsOptional  bool   `json:"is_optional"`
}

type FieldInput struct {
	FieldName        string                 `json:"field_name"`
	FieldType        string                 `json:"field_type"`
	StepNumber       int                    `json:"step_number"`
	Label            map[string]string      `json:"label"`
	Placeholder      map[string]string      `json:"placeholder"`
	HelpText         map[string]string      `json:"help_text"`
	ValidationRules  map[string]interface{} `json:"validation_rules"`
	Options          map[string]interface{} `json:"options"`
	DisplayOrder     int                    `json:"display_order"`
	IsRequired       bool                   `json:"is_required"`
	IsReadonly       bool                   `json:"is_readonly"`
	DefaultValue     *string                `json:"default_value"`
	ConditionalLogic map[string]interface{} `json:"conditional_logic"`
	FileConfig       map[string]interface{} `json:"file_config"`
}

type PersistenceConfigInput struct {
	PersistenceMode     string                   `json:"persistence_mode"`
	TargetConfigs       []map[string]interface{} `json:"target_configs"`
	FieldMappings       map[string]interface{}   `json:"field_mappings"`
	TransformationRules map[string]interface{}   `json:"transformation_rules"`
	ValidationHooks     []map[string]interface{} `json:"validation_hooks"`
}

// FormSubmissionInput for processing submissions
type FormSubmissionInput struct {
	FormDefinitionID uuid.UUID                 `json:"form_definition_id"`
	UserID           uuid.UUID                 `json:"user_id"`
	Status           string                    `json:"status"`
	Data             map[string]interface{}    `json:"data"`
	Files            []FormSubmissionFileInput `json:"files"`
	Metadata         map[string]interface{}    `json:"metadata"`
}

type FormSubmissionFileInput struct {
	FieldName       string `json:"field_name"`
	FileName        string `json:"file_name"`
	FilePath        string `json:"file_path"`
	FileSize        int64  `json:"file_size"`
	MimeType        string `json:"mime_type"`
	Bucket          string `json:"bucket"`
	StorageProvider string `json:"storage_provider"`
}

// FieldMapping and TargetConfig for persistence
type FieldMapping struct {
	FormField   string  `json:"form_field"`
	TableName   string  `json:"table_name"`
	ColumnName  string  `json:"column_name"`
	DataType    string  `json:"data_type"`
	Transform   *string `json:"transform,omitempty"`
	IsEncrypted bool    `json:"is_encrypted"`
	MetaKey     string  `json:"meta_key,omitempty"` // For meta tables
}

type TargetConfig struct {
	TableName  string                 `json:"table_name"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Priority   int                    `json:"priority"`
}

type SaveStepProgressInput struct {
	SubmissionID uuid.UUID              `json:"submission_id"`
	StepID       uuid.UUID              `json:"step_id"`
	StepNumber   int32                  `json:"step_number"`
	Status       string                 `json:"status"`
	Data         map[string]interface{} `json:"data"`
	UserID       uuid.UUID              `json:"user_id"`
}
