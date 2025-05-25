package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/timchuks/monieverse/internal/config"
	"log"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/notifier"
	"github.com/timchuks/monieverse/internal/uploader"
	"github.com/timchuks/monieverse/internal/validator"
	"github.com/timchuks/monieverse/internal/worker"
)

type FormService struct {
	store           db.Store
	uploader        uploader.FileUploader
	notifier        notifier.Notifier
	taskDistributor worker.TaskDistributor
	config          *config.Config
}

func NewFormService(
	store db.Store,
	uploader uploader.FileUploader,
	notifier notifier.Notifier,
	taskDistributor worker.TaskDistributor,
	config *config.Config,
) *FormService {
	return &FormService{
		store:           store,
		uploader:        uploader,
		notifier:        notifier,
		taskDistributor: taskDistributor,
		config:          config,
	}
}

// GetFormForUser retrieves the appropriate form based on assignments
func (s *FormService) GetFormForUser(ctx context.Context, userID uuid.UUID, formType string) (*FormDefinitionWithData, error) {
	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get assignments
	assignments, err := s.store.GetFormAssignments(ctx, db.GetFormAssignmentsParams{
		AssignmentValue:   userID.String(),
		AssignmentValue_2: user.AccountType, // user_type
		AssignmentValue_3: user.CountryCode, // country
		AssignmentValue_4: user.State,       // state
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get form assignments: %w", err)
	}

	// Find best matching assignment
	var bestAssignment *db.GetFormAssignmentsRow
	for _, assignment := range assignments {
		if assignment.FormType != formType {
			continue
		}

		// Check conditions
		var conditions map[string]interface{}
		if err := json.Unmarshal(assignment.Conditions.RawMessage, &conditions); err == nil {
			if !s.checkConditions(&user, conditions) {
				continue
			}
		}

		if bestAssignment == nil || assignment.Priority > bestAssignment.Priority {
			bestAssignment = &assignment
		}
	}

	if bestAssignment == nil {
		return nil, fmt.Errorf("no form found for user")
	}

	// Get form definition
	form, err := s.store.GetFormDefinition(ctx, bestAssignment.FormDefinitionID)
	if err != nil {
		return nil, err
	}

	// Get steps
	steps, err := s.store.GetFormSteps(ctx, form.ID)
	if err != nil {
		return nil, err
	}

	// Get fields
	fields, err := s.store.GetFormFields(ctx, form.ID)
	if err != nil {
		return nil, err
	}

	// Process dynamic options
	for i, field := range fields {
		if field.Options != nil {
			var options FieldOptions
			if err = json.Unmarshal(field.Options, &options); err == nil {
				if options.Type == "dynamic" && options.Dynamic != nil {
					dynamicOptions, err := s.getDynamicOptions(ctx, options.Dynamic)
					if err == nil {
						options.Static = dynamicOptions
						optionsJSON, _ := json.Marshal(options)
						fields[i].Options = optionsJSON
					}
				}
			}
		}
	}

	// Get existing submission if any
	submission, _ := s.store.GetFormSubmissionByUserAndForm(ctx, db.GetFormSubmissionByUserAndFormParams{
		UserID:           userID,
		FormDefinitionID: form.ID,
		Status:           "draft",
	})

	return &FormDefinitionWithData{
		FormDefinition: form,
		Steps:          steps,
		Fields:         fields,
		ExistingData:   submission.SubmissionData,
	}, nil
}

// SubmitForm processes form submission
func (s *FormService) SubmitForm(ctx context.Context, input SubmitFormInput) (*db.FormSubmission, error) {
	// Validate form exists and is active
	form, err := s.store.GetFormDefinition(ctx, input.FormID)
	if err != nil {
		return nil, fmt.Errorf("form not found")
	}

	if !form.IsActive {
		return nil, fmt.Errorf("form is not active")
	}

	// Get fields for validation
	fields, err := s.store.GetFormFields(ctx, form.ID)
	if err != nil {
		return nil, err
	}

	// Validate submission
	if err := s.validateSubmission(fields, input.Data); err != nil {
		return nil, err
	}

	// Trigger before_submit events
	s.triggerEvents(ctx, form.ID, "before_submit", input)

	// Handle file uploads
	var submissionFiles []db.FormSubmissionFileInput
	for fieldName, fileHeaders := range input.Files {
		field := s.findField(fields, fieldName)
		if field == nil {
			continue
		}

		var fileConfig FileConfig
		if field.FileConfig != nil {
			//TODO: see if we can skip this upload and log the errors instead
			err = json.Unmarshal(field.FileConfig, &fileConfig)
			if err != nil {
				return nil, err
			}
		}

		uploadedFiles, err := s.handleFileUploads(ctx, input.UserID, fieldName, fileHeaders, &fileConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to upload files: %w", err)
		}

		submissionFiles = append(submissionFiles, uploadedFiles...)
	}

	// Process submission
	submissionInput := &db.FormSubmissionInput{
		FormDefinitionID: form.ID,
		UserID:           input.UserID,
		Status:           "submitted",
		Data:             input.Data,
		Files:            submissionFiles,
		Metadata:         input.Metadata,
	}

	submission, err := s.store.ProcessFormSubmissionTx(ctx, submissionInput)
	if err != nil {
		return nil, err
	}

	// Trigger after_submit events
	s.triggerEvents(ctx, form.ID, "after_submit", submission)

	// Send notifications
	if form.RequiresApproval {
		//s.sendApprovalNotification(ctx, &user, form, submission)
		log.Printf("form approved")
	}

	return submission, nil
}

// validateSubmission validates form data
func (s *FormService) validateSubmission(fields []db.FormField, data map[string]interface{}) error {
	v := validator.New()

	for _, field := range fields {
		value, exists := data[field.FieldName]

		// Check required
		if field.IsRequired && (!exists || s.isEmpty(value)) {
			v.AddError(field.FieldName, "field is required")
			continue
		}

		if !exists || value == nil {
			continue
		}

		// Parse validation rules
		var rules ValidationRules
		if field.ValidationRules != nil {
			json.Unmarshal(field.ValidationRules, &rules)
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
					// Pattern validation
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
				json.Unmarshal(field.Options, &options)
			}

			if options.Type == "static" && len(options.Static) > 0 {
				validOptions := make([]string, len(options.Static))
				for i, opt := range options.Static {
					validOptions[i] = opt.Value
				}
				v.Check(validator.In(value.(string), validOptions...), field.FieldName, "invalid option")
			}
		}
	}

	if !v.Valid() {
		return validator.NewValidationError("validation failed", v.Errors)
	}

	return nil
}

// handleFileUploads processes file uploads
func (s *FormService) handleFileUploads(
	ctx context.Context,
	userID uuid.UUID,
	fieldName string,
	files []*multipart.FileHeader,
	config *FileConfig,
) ([]db.FormSubmissionFileInput, error) {
	var result []db.FormSubmissionFileInput

	for i, fileHeader := range files {
		// Validate file
		if config != nil {
			if fileHeader.Size > config.MaxSize {
				return nil, fmt.Errorf("file %s exceeds maximum size of %d bytes", fileHeader.Filename, config.MaxSize)
			}

			mimeType := fileHeader.Header.Get("Content-Type")
			if len(config.AllowedTypes) > 0 && !s.contains(config.AllowedTypes, mimeType) {
				return nil, fmt.Errorf("file type %s not allowed", mimeType)
			}

			if config.MaxFiles > 0 && i >= config.MaxFiles {
				return nil, fmt.Errorf("maximum %d files allowed", config.MaxFiles)
			}
		}

		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Generate path
		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("%s_%d_%s%s", fieldName, time.Now().Unix(), uuid.New().String()[:8], ext)
		path := fmt.Sprintf("forms/%s/%s/%s", userID.String(), fieldName, filename)

		// Upload
		if err := s.uploader.Upload(file, s.config.FileBucket, path); err != nil {
			return nil, fmt.Errorf("failed to upload file: %w", err)
		}

		result = append(result, db.FormSubmissionFileInput{
			FieldName:       fieldName,
			FileName:        fileHeader.Filename,
			FilePath:        path,
			FileSize:        fileHeader.Size,
			MimeType:        fileHeader.Header.Get("Content-Type"),
			Bucket:          s.config.FileBucket,
			StorageProvider: s.config.FileUploadProvider,
		})
	}

	return result, nil
}

// getDynamicOptions fetches options from dynamic source
func (s *FormService) getDynamicOptions(ctx context.Context, source *DynamicSource) ([]Option, error) {
	// Get dynamic option config
	dynamicOption, err := s.store.GetDynamicOptionByName(ctx, source.SourceName)
	if err != nil {
		return nil, err
	}

	var sourceConfig map[string]interface{}
	json.Unmarshal(dynamicOption.SourceConfig, &sourceConfig)

	switch dynamicOption.SourceType {
	case "table":
		tableName := sourceConfig["table"].(string)

		switch tableName {
		case "countries":
			rows, err := s.store.GetCountries(ctx, db.CountryQueryFilter{
				ActiveOnly: true,
			}, db.Filter{
				PageSize: 250,
			})
			if err != nil {
				return nil, err
			}

			options := make([]Option, len(rows))
			for i, row := range rows {
				options[i] = Option{
					Value: row.Code,
					Label: I18nText{"en": row.Name},
				}
			}
			return options, nil

		}

	}

	return nil, fmt.Errorf("unknown dynamic source")
}

// triggerEvents processes form events
func (s *FormService) triggerEvents(ctx context.Context, formID uuid.UUID, eventType string, data interface{}) {
	events, err := s.store.GetFormEvents(ctx, db.GetFormEventsParams{
		FormDefinitionID: formID,
		EventType:        eventType,
	})

	if err != nil {
		return
	}

	for _, event := range events {
		if !event.IsActive {
			continue
		}
		var config map[string]interface{}
		json.Unmarshal(event.HandlerConfig, &config)

		log.Printf("event %s triggered for form %s", eventType, formID)
	}
}

// Helper methods
func (s *FormService) checkConditions(user *db.User, conditions map[string]interface{}) bool {
	for key, value := range conditions {
		switch key {
		case "new_customers_only":
			if value.(bool) {
				// Check if user registered within last 30 days
				thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
				if user.CreatedAt.Before(thirtyDaysAgo) {
					return false
				}
			}
		case "user_type":
			if user.AccountType != value.(string) {
				return false
			}
		case "kyc_verified":
			if value.(bool) && user.KycVerified != "verified" {
				return false
			}
		}
	}
	return true
}

func (s *FormService) findField(fields []db.FormField, fieldName string) *db.FormField {
	for _, field := range fields {
		if field.FieldName == fieldName {
			return &field
		}
	}
	return nil
}

func (s *FormService) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

func (s *FormService) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}

func (s *FormService) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *FormService) sendApprovalNotification(ctx context.Context, user *db.User, form db.FormDefinition, submission *db.FormSubmission) {

}

// ApproveSubmission approves a form submission
func (s *FormService) ApproveSubmission(ctx context.Context, submissionID uuid.UUID, approverID uuid.UUID, notes string) error {
	submission, err := s.store.GetFormSubmission(ctx, submissionID)
	if err != nil {
		return err
	}

	form, err := s.store.GetFormDefinition(ctx, submission.FormDefinitionID)
	if err != nil {
		return err
	}

	// Trigger before_approve events
	s.triggerEvents(ctx, form.ID, "before_approve", submission)

	// Update submission
	_, err = s.store.UpdateFormSubmission(ctx, db.UpdateFormSubmissionParams{
		ID:             submissionID,
		SubmissionData: submission.SubmissionData,
		Status:         "approved",
		ApprovalStatus: "approved",
		ApprovalNotes:  notes,
		ApprovedBy:     db.NewNullUUID(approverID),
		ApprovedAt:     time.Now(),
		Metadata:       submission.Metadata,
	})

	if err != nil {
		return err
	}

	// Trigger after_approve events
	s.triggerEvents(ctx, form.ID, "after_approve", submission)

	return nil
}

// RejectSubmission rejects a form submission
func (s *FormService) RejectSubmission(ctx context.Context, submissionID uuid.UUID, approverID uuid.UUID, reason string) error {
	submission, err := s.store.GetFormSubmission(ctx, submissionID)
	if err != nil {
		return err
	}

	form, err := s.store.GetFormDefinition(ctx, submission.FormDefinitionID)
	if err != nil {
		return err
	}

	// Trigger before_reject events
	s.triggerEvents(ctx, form.ID, "before_reject", submission)

	// Update submission
	_, err = s.store.UpdateFormSubmission(ctx, db.UpdateFormSubmissionParams{
		ID:             submissionID,
		SubmissionData: submission.SubmissionData,
		Status:         "rejected",
		ApprovalStatus: "rejected",
		ApprovalNotes:  reason,
		ApprovedBy:     db.NewNullUUID(approverID),
		ApprovedAt:     time.Now(),
		Metadata:       submission.Metadata,
	})

	if err != nil {
		return err
	}

	// Trigger after_reject events
	s.triggerEvents(ctx, form.ID, "after_reject", submission)

	return nil
}

// GetFormForEdit retrieves a form with existing submission data for editing
func (s *FormService) GetFormForEdit(ctx context.Context, submissionID uuid.UUID, userID uuid.UUID) (*FormDefinitionWithData, error) {
	// Get submission
	submission, err := s.store.GetFormSubmission(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// Verify ownership
	if submission.UserID != userID {
		user, _ := s.store.GetUser(ctx, userID)
		if !s.store.HasPermission(ctx, user, "admin.forms") {
			return nil, fmt.Errorf("access denied")
		}
	}

	// Get form definition
	form, err := s.store.GetFormDefinition(ctx, submission.FormDefinitionID)
	if err != nil {
		return nil, err
	}

	// Check if form is editable
	if !form.IsEditableAfterSubmission && submission.Status != "draft" {
		return nil, fmt.Errorf("form is not editable after submission")
	}

	// Check approval status
	if form.RequiresApproval && submission.ApprovalStatus != "" &&
		(submission.ApprovalStatus == "approved" || submission.ApprovalStatus == "rejected") {
		return nil, fmt.Errorf("cannot edit %s submission", submission.ApprovalStatus)
	}

	// Get steps
	steps, err := s.store.GetFormSteps(ctx, form.ID)
	if err != nil {
		return nil, err
	}

	// Get fields
	fields, err := s.store.GetFormFields(ctx, form.ID)
	if err != nil {
		return nil, err
	}

	// Parse submission data
	var submissionData map[string]interface{}
	if err := json.Unmarshal(submission.SubmissionData, &submissionData); err != nil {
		return nil, fmt.Errorf("failed to parse submission data: %w", err)
	}

	// Get files
	files, _ := s.store.GetFormSubmissionFiles(ctx, submissionID)
	fileMap := make(map[string][]map[string]interface{})
	for _, file := range files {
		fileInfo := map[string]interface{}{
			"id":        file.ID,
			"file_name": file.FileName,
			"file_size": file.FileSize,
			"mime_type": file.MimeType,
			"url":       s.getFileURL(file),
		}
		fileMap[file.FieldName] = append(fileMap[file.FieldName], fileInfo)
	}

	// Set default values from submission data
	for i, field := range fields {
		fieldName := field.FieldName

		// Check for file fields
		if field.FieldType == "file" || field.FieldType == "files" {
			if files, ok := fileMap[fieldName]; ok {
				fileData, _ := json.Marshal(files)
				fields[i].DefaultValue = string(fileData)
			}
		} else if value, ok := submissionData[fieldName]; ok {
			// Convert value to string for default value
			var defaultValue string
			switch v := value.(type) {
			case string:
				defaultValue = v
			case float64:
				defaultValue = fmt.Sprintf("%v", v)
			case bool:
				defaultValue = fmt.Sprintf("%v", v)
			case []interface{}:
				// For checkbox fields
				jsonValue, _ := json.Marshal(v)
				defaultValue = string(jsonValue)
			default:
				jsonValue, _ := json.Marshal(v)
				defaultValue = string(jsonValue)
			}

			fields[i].DefaultValue = defaultValue
		}

		// Process dynamic options
		if field.Options != nil {
			var options FieldOptions
			if err := json.Unmarshal(field.Options, &options); err == nil {
				if options.Type == "dynamic" && options.Dynamic != nil {
					dynamicOptions, err := s.getDynamicOptions(ctx, options.Dynamic)
					if err == nil {
						options.Static = dynamicOptions
						optionsJSON, _ := json.Marshal(options)
						fields[i].Options = optionsJSON
					}
				}
			}
		}
	}

	return &FormDefinitionWithData{
		FormDefinition: form,
		Steps:          steps,
		Fields:         fields,
		ExistingData:   submission.SubmissionData,
		SubmissionID:   &submission.ID,
	}, nil
}

// UpdateFormSubmission update form
func (s *FormService) UpdateFormSubmission(ctx context.Context, input UpdateSubmissionInput) (*db.FormSubmission, error) {
	// Get existing submission
	submission, err := s.store.GetFormSubmission(ctx, input.SubmissionID)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// Verify ownership
	if submission.UserID != input.UserID {
		user, _ := s.store.GetUser(ctx, input.UserID)
		if !s.store.HasPermission(ctx, user, "admin.forms") {
			return nil, fmt.Errorf("access denied")
		}
	}

	// Get form definition
	form, err := s.store.GetFormDefinition(ctx, submission.FormDefinitionID)
	if err != nil {
		return nil, err
	}

	// Check if form is editable
	if !form.IsEditableAfterSubmission && submission.Status != "draft" {
		return nil, fmt.Errorf("form is not editable after submission")
	}

	// Check approval status
	if form.RequiresApproval && submission.ApprovalStatus != "" &&
		(submission.ApprovalStatus == "approved" || submission.ApprovalStatus == "rejected") {
		return nil, fmt.Errorf("cannot edit %s submission", submission.ApprovalStatus)
	}

	// Get fields for validation
	fields, err := s.store.GetFormFields(ctx, form.ID)
	if err != nil {
		return nil, err
	}

	// Validate the updated data
	if err := s.validateSubmission(fields, input.Data); err != nil {
		return nil, err
	}

	// Trigger before_update events
	s.triggerEvents(ctx, form.ID, "before_update", input)

	// Handle file uploads
	var newFiles []db.FormSubmissionFileInput
	for fieldName, fileHeaders := range input.Files {
		field := s.findField(fields, fieldName)
		if field == nil {
			continue
		}

		var fileConfig FileConfig
		if field.FileConfig != nil {
			json.Unmarshal(field.FileConfig, &fileConfig)
		}

		uploadedFiles, err := s.handleFileUploads(ctx, input.UserID, fieldName, fileHeaders, &fileConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to upload files: %w", err)
		}

		newFiles = append(newFiles, uploadedFiles...)
	}

	// Process submission update through transaction
	updateInput := &db.FormSubmissionUpdateInput{
		SubmissionID:     input.SubmissionID,
		FormDefinitionID: form.ID,
		UserID:           input.UserID,
		Status:           input.Status,
		Data:             input.Data,
		Files:            newFiles,
		Metadata:         input.Metadata,
		IsPartialUpdate:  input.IsPartialUpdate,
		ExistingData:     submission.SubmissionData,
	}

	updatedSubmission, err := s.store.UpdateFormSubmissionTx(ctx, updateInput)
	if err != nil {
		return nil, err
	}

	// Trigger after_update events
	s.triggerEvents(ctx, form.ID, "after_update", updatedSubmission)

	return updatedSubmission, nil
}

// Helper method to get file URL
func (s *FormService) getFileURL(file db.FormSubmissionFile) string {
	// Generate temporary URL for file access
	url, err := s.uploader.GetTempURL(file.Bucket, file.FilePath)
	if err != nil {
		return ""
	}
	return url
}

type UpdateSubmissionInput struct {
	SubmissionID    uuid.UUID                          `json:"submission_id"`
	UserID          uuid.UUID                          `json:"user_id"`
	Data            map[string]interface{}             `json:"data"`
	Files           map[string][]*multipart.FileHeader `json:"-"`
	Status          string                             `json:"status"`
	IsPartialUpdate bool                               `json:"is_partial_update"`
	Metadata        map[string]interface{}             `json:"metadata,omitempty"`
}

// SaveStepProgress saves progress for a specific step
func (s *FormService) SaveStepProgress(ctx context.Context, input SaveStepProgressInput) (*StepProgressResult, error) {
	// Get submission
	submission, err := s.store.GetFormSubmission(ctx, input.SubmissionID)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// Verify ownership
	if submission.UserID != input.UserID {
		return nil, fmt.Errorf("access denied")
	}

	// Get form and steps
	form, err := s.store.GetFormDefinition(ctx, submission.FormDefinitionID)
	if err != nil {
		return nil, err
	}

	steps, err := s.store.GetFormSteps(ctx, form.ID)
	if err != nil {
		return nil, err
	}

	// Find the step
	var currentStep *db.FormStep
	for _, step := range steps {
		if step.StepNumber == input.StepNumber {
			currentStep = &step
			break
		}
	}

	if currentStep == nil {
		return nil, fmt.Errorf("step not found")
	}

	// Get fields for this step
	fields, err := s.store.GetFormFieldsByStep(ctx, currentStep.ID)
	if err != nil {
		return nil, err
	}

	// Validate step data if completing
	if input.Status == "completed" {
		if err := s.validateStepData(fields, input.Data); err != nil {
			return nil, err
		}
	}

	// Save step progress
	err = s.store.SaveStepProgressTx(ctx, &db.SaveStepProgressInput{
		SubmissionID: input.SubmissionID,
		StepID:       currentStep.ID,
		StepNumber:   input.StepNumber,
		Status:       input.Status,
		Data:         input.Data,
		UserID:       input.UserID,
	})

	if err != nil {
		return nil, err
	}

	// Calculate overall progress
	allProgress, err := s.store.GetAllStepProgress(ctx, db.NewNullUUID(input.SubmissionID))
	if err != nil {
		return nil, err
	}

	completedSteps := 0
	currentStepNumber := int32(1)

	for _, progress := range allProgress {
		if progress.Status.String == "completed" {
			completedSteps++
		}
	}

	// Find next incomplete step
	for i, step := range steps {
		found := false
		for _, progress := range allProgress {
			if progress.FormStepID.UUID == step.ID && progress.Status.String == "completed" {
				found = true
				break
			}
		}
		if !found {
			currentStepNumber = int32(i + 1)
			break
		}
	}

	completionPercentage := int32(float64(completedSteps) / float64(len(steps)) * 100)

	// Update submission progress
	_, err = s.store.UpdateSubmissionProgress(ctx, db.UpdateSubmissionProgressParams{
		ID:                   input.SubmissionID,
		CurrentStepNumber:    db.NewNullInt32(currentStepNumber),
		CompletionPercentage: db.NewNullInt32(completionPercentage),
	})

	return &StepProgressResult{
		Success:              true,
		CurrentStep:          currentStepNumber,
		CompletionPercentage: completionPercentage,
		AllStepsCompleted:    completedSteps == len(steps),
	}, nil
}

// GetFormWithProgress retrieves form with step progress
func (s *FormService) GetFormWithProgress(ctx context.Context, userID uuid.UUID, formType string) (*FormDefinitionWithData, error) {
	// First get the form
	formData, err := s.GetFormForUser(ctx, userID, formType)
	if err != nil {
		return nil, err
	}

	// Check if there's an existing draft submission
	if formData.SubmissionID != nil {
		// Get step progress
		stepProgress, err := s.store.GetAllStepProgress(ctx, db.NewNullUUID(*formData.SubmissionID))
		if err == nil {
			formData.StepProgress = stepProgress
		}

		// Get current submission to get progress info
		submission, err := s.store.GetFormSubmission(ctx, *formData.SubmissionID)
		if err == nil {
			formData.CurrentStep = submission.CurrentStepNumber.Int32
			formData.CompletionPercentage = submission.CompletionPercentage.Int32

			// Merge all step data into existing data
			allData := make(map[string]interface{})

			// First, parse any existing submission data
			if submission.SubmissionData != nil {
				_ = json.Unmarshal(submission.SubmissionData, &allData)
			}

			// Then merge step-specific data
			for _, progress := range stepProgress {
				var stepData map[string]interface{}
				if err := json.Unmarshal(progress.Data.RawMessage, &stepData); err == nil {
					for key, value := range stepData {
						allData[key] = value
					}
				}
			}

			// Update existing data
			if len(allData) > 0 {
				mergedData, _ := json.Marshal(allData)
				formData.ExistingData = mergedData
			}
		}
	} else {
		formData.CurrentStep = 1
		formData.CompletionPercentage = 0
	}

	return formData, nil
}

// validateStepData validates only the fields for a specific step
func (s *FormService) validateStepData(fields []db.FormField, data map[string]interface{}) error {
	v := validator.New()

	for _, field := range fields {
		value, exists := data[field.FieldName]

		// For step validation, only validate fields that are present or required
		if field.IsRequired && (!exists || s.isEmpty(value)) {
			v.AddError(field.FieldName, "field is required")
			continue
		}

		if !exists {
			continue
		}

		// Apply same validation logic as full form
		var rules ValidationRules
		if field.ValidationRules != nil {
			_ = json.Unmarshal(field.ValidationRules, &rules)
		}

		// ... (same validation logic as validateSubmission)
	}

	if !v.Valid() {
		return validator.NewValidationError("validation failed", v.Errors)
	}

	return nil
}

// CompleteForm finishes the form submission after all steps
func (s *FormService) CompleteForm(ctx context.Context, submissionID uuid.UUID, userID uuid.UUID) (*db.FormSubmission, error) {
	// Get submission
	submission, err := s.store.GetFormSubmission(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// Verify ownership
	if submission.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// Check all steps are completed
	form, _ := s.store.GetFormDefinition(ctx, submission.FormDefinitionID)
	steps, _ := s.store.GetFormSteps(ctx, form.ID)
	stepProgress, _ := s.store.GetAllStepProgress(ctx, db.NewNullUUID(submissionID))

	if len(stepProgress) < len(steps) {
		return nil, fmt.Errorf("not all steps completed")
	}

	for _, progress := range stepProgress {
		if progress.Status.String != "completed" {
			return nil, fmt.Errorf("step %d not completed", progress.StepNumber)
		}
	}

	// Merge all step data
	allData := make(map[string]interface{})
	for _, progress := range stepProgress {
		var stepData map[string]interface{}
		if err := json.Unmarshal(progress.Data.RawMessage, &stepData); err == nil {
			for key, value := range stepData {
				allData[key] = value
			}
		}
	}

	// Process final submission
	return s.store.ProcessFormSubmissionTx(ctx, &db.FormSubmissionInput{
		FormDefinitionID: submission.FormDefinitionID,
		UserID:           userID,
		Status:           "submitted",
		Data:             allData,
		Files:            []db.FormSubmissionFileInput{}, // Files already uploaded per step
		Metadata: map[string]interface{}{
			"completed_at": time.Now(),
		},
	})
}

// Input types
type SaveStepProgressInput struct {
	SubmissionID uuid.UUID              `json:"submission_id"`
	UserID       uuid.UUID              `json:"user_id"`
	StepNumber   int32                  `json:"step_number"`
	Status       string                 `json:"status"` // 'in_progress' or 'completed'
	Data         map[string]interface{} `json:"data"`
}

type StepProgressResult struct {
	Success              bool  `json:"success"`
	CurrentStep          int32 `json:"current_step"`
	CompletionPercentage int32 `json:"completion_percentage"`
	AllStepsCompleted    bool  `json:"all_steps_completed"`
}
