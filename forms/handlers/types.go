package handlers

import (
	"time"
)

// CreateFormRequest for creating forms
type CreateFormRequest struct {
	Name                      string                  `json:"name" binding:"required"`
	Slug                      string                  `json:"slug" binding:"required"`
	Description               string                  `json:"description"`
	FormType                  string                  `json:"form_type" binding:"required"`
	IsMultiStep               bool                    `json:"is_multi_step"`
	RequiresApproval          bool                    `json:"requires_approval"`
	IsEditableAfterSubmission bool                    `json:"is_editable_after_submission"`
	ApprovalWorkflow          *ApprovalWorkflowInput  `json:"approval_workflow,omitempty"`
	Steps                     []StepInput             `json:"steps,omitempty"`
	Fields                    []FieldInput            `json:"fields" binding:"required"`
	PersistenceConfig         *PersistenceConfigInput `json:"persistence_config,omitempty"`
}

type ApprovalWorkflowInput struct {
	States      []ApprovalStateInput `json:"states"`
	Transitions []TransitionInput    `json:"transitions"`
}

type ApprovalStateInput struct {
	Name        string            `json:"name"`
	Label       map[string]string `json:"label"`
	IsFinal     bool              `json:"is_final"`
	Permissions []string          `json:"permissions"`
}

type TransitionInput struct {
	From        string            `json:"from"`
	To          string            `json:"to"`
	Label       map[string]string `json:"label"`
	Permissions []string          `json:"permissions"`
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
	StepNumber       *int                   `json:"step_number,omitempty"`
	Label            map[string]string      `json:"label"`
	Placeholder      map[string]string      `json:"placeholder,omitempty"`
	HelpText         map[string]string      `json:"help_text,omitempty"`
	ValidationRules  map[string]interface{} `json:"validation_rules"`
	Options          *OptionsInput          `json:"options,omitempty"`
	DisplayOrder     int                    `json:"display_order"`
	IsRequired       bool                   `json:"is_required"`
	IsReadonly       bool                   `json:"is_readonly"`
	DefaultValue     *string                `json:"default_value,omitempty"`
	ConditionalLogic *ConditionalLogicInput `json:"conditional_logic,omitempty"`
	FileConfig       *FileConfigInput       `json:"file_config,omitempty"`
}

type OptionsInput struct {
	Type    string              `json:"type"`
	Static  []OptionInput       `json:"static,omitempty"`
	Dynamic *DynamicSourceInput `json:"dynamic,omitempty"`
}

type OptionInput struct {
	Value string            `json:"value"`
	Label map[string]string `json:"label"`
}

type DynamicSourceInput struct {
	SourceName   string            `json:"source_name"`
	FilterParams map[string]string `json:"filter_params,omitempty"`
}

type ConditionalLogicInput struct {
	Action     string           `json:"action"`
	Conditions []ConditionInput `json:"conditions"`
	Logic      string           `json:"logic"`
}

type ConditionInput struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type FileConfigInput struct {
	MaxSize      int64    `json:"max_size"`
	AllowedTypes []string `json:"allowed_types"`
	MaxFiles     int      `json:"max_files,omitempty"`
}

type PersistenceConfigInput struct {
	PersistenceMode     string                       `json:"persistence_mode"`
	TargetConfigs       []TargetConfigInput          `json:"target_configs"`
	FieldMappings       map[string]FieldMappingInput `json:"field_mappings"`
	TransformationRules map[string]interface{}       `json:"transformation_rules,omitempty"`
	ValidationHooks     []ValidationHookInput        `json:"validation_hooks,omitempty"`
}

type TargetConfigInput struct {
	TableName  string                 `json:"table_name"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Priority   int                    `json:"priority"`
}

type FieldMappingInput struct {
	FormField   string  `json:"form_field"`
	TableName   string  `json:"table_name"`
	ColumnName  string  `json:"column_name"`
	DataType    string  `json:"data_type"`
	Transform   *string `json:"transform,omitempty"`
	IsEncrypted bool    `json:"is_encrypted"`
	MetaKey     string  `json:"meta_key,omitempty"`
}

type ValidationHookInput struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// UpdateFormRequest for updating forms
type UpdateFormRequest struct {
	Name                      string                 `json:"name"`
	Description               string                 `json:"description"`
	IsActive                  bool                   `json:"is_active"`
	IsEditableAfterSubmission bool                   `json:"is_editable_after_submission"`
	ApprovalWorkflow          *ApprovalWorkflowInput `json:"approval_workflow,omitempty"`
}

// CreateAssignmentRequest for form assignments
type CreateAssignmentRequest struct {
	AssignmentType  string                 `json:"assignment_type" binding:"required"`
	AssignmentValue string                 `json:"assignment_value" binding:"required"`
	Conditions      map[string]interface{} `json:"conditions,omitempty"`
	Priority        int                    `json:"priority"`
	ValidFrom       *time.Time             `json:"valid_from,omitempty"`
	ValidUntil      *time.Time             `json:"valid_until,omitempty"`
}

// PersistenceConfigRequest for persistence configuration
type PersistenceConfigRequest struct {
	PersistenceMode     string                       `json:"persistence_mode" binding:"required"`
	TargetConfigs       []TargetConfigInput          `json:"target_configs" binding:"required"`
	FieldMappings       map[string]FieldMappingInput `json:"field_mappings" binding:"required"`
	TransformationRules map[string]interface{}       `json:"transformation_rules,omitempty"`
	ValidationHooks     []ValidationHookInput        `json:"validation_hooks,omitempty"`
}

// ApprovalRequest for approve/reject
type ApprovalRequest struct {
	Notes  string `json:"notes"`
	Reason string `json:"reason"`
}
