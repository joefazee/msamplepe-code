package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

// CreateFormDefinitionTx creates a form with all its fields and steps in a transaction
func (store *SQLStore) CreateFormDefinitionTx(ctx context.Context, input *FormDefinitionInput) (*FormDefinition, error) {
	var form FormDefinition

	err := store.execTx(ctx, func(q *Queries) error {
		// Create form definition
		formParams := CreateFormDefinitionParams{
			ID:               uuid.New(),
			Name:             input.Name,
			Slug:             input.Slug,
			Description:      input.Description,
			FormType:         input.FormType,
			Version:          1,
			IsActive:         true,
			IsMultiStep:      input.IsMultiStep,
			RequiresApproval: input.RequiresApproval,
			CreatedBy:        NewNullUUID(input.CreatedBy),
		}

		if input.ApprovalWorkflow != nil {
			workflowJSON, err := json.Marshal(input.ApprovalWorkflow)
			if err != nil {
				return err
			}
			formParams.ApprovalWorkflow = workflowJSON
		} else {
			formParams.ApprovalWorkflow = json.RawMessage("{}")
		}

		createdForm, err := q.CreateFormDefinition(ctx, formParams)
		if err != nil {
			return err
		}
		form = createdForm

		// Create steps if multi-step
		stepMap := make(map[int]uuid.UUID)
		if input.IsMultiStep && len(input.Steps) > 0 {
			for _, step := range input.Steps {
				stepID := uuid.New()
				stepMap[step.StepNumber] = stepID

				_, err := q.CreateFormStep(ctx, CreateFormStepParams{
					ID:               stepID,
					FormDefinitionID: form.ID,
					StepNumber:       int32(step.StepNumber),
					Name:             step.Name,
					Description:      step.Description,
					IsOptional:       step.IsOptional,
				})
				if err != nil {
					return err
				}
			}
		} else {
			stepNum := int32(1)
			stepID := uuid.New()
			stepMap[int(stepNum)] = stepID
			_, err := q.CreateFormStep(ctx, CreateFormStepParams{
				ID:               stepID,
				FormDefinitionID: form.ID,
				StepNumber:       stepNum,
				Name:             input.Name,
				Description:      input.Description,
				IsOptional:       false,
			})
			if err != nil {
				return err
			}
		}

		// Create fields
		for _, field := range input.Fields {
			fieldParams := CreateFormFieldParams{
				ID:               uuid.New(),
				FormDefinitionID: form.ID,
				FieldName:        field.FieldName,
				FieldType:        field.FieldType,
				DisplayOrder:     int32(field.DisplayOrder),
				IsRequired:       field.IsRequired,
				IsReadonly:       field.IsReadonly,
			}

			// Handle step assignment
			if field.StepNumber > 0 {
				if stepID, ok := stepMap[field.StepNumber]; ok {
					fieldParams.FormStepID = stepID
				}
			}

			// Marshal JSON fields
			if field.Label != nil {
				labelJSON, err := json.Marshal(field.Label)
				if err != nil {
					return err
				}
				fieldParams.Label = labelJSON
			} else {
				fieldParams.Label = json.RawMessage("{}")
			}

			if field.Placeholder != nil {
				placeholderJSON, err := json.Marshal(field.Placeholder)
				if err != nil {
					return err
				}
				fieldParams.Placeholder = placeholderJSON
			} else {
				fieldParams.Placeholder = json.RawMessage("{}")
			}

			if field.HelpText != nil {
				helpTextJSON, err := json.Marshal(field.HelpText)
				if err != nil {
					return err
				}
				fieldParams.HelpText = helpTextJSON
			} else {
				fieldParams.HelpText = json.RawMessage("{}")
			}

			if field.ValidationRules != nil {
				rulesJSON, err := json.Marshal(field.ValidationRules)
				if err != nil {
					return err
				}
				fieldParams.ValidationRules = rulesJSON
			} else {
				fieldParams.ValidationRules = json.RawMessage("{}")
			}

			if field.Options != nil {
				optionsJSON, err := json.Marshal(field.Options)
				if err != nil {
					return err
				}
				fieldParams.Options = optionsJSON
			} else {
				fieldParams.Options = json.RawMessage("{}")
			}

			if field.DefaultValue != nil {
				fieldParams.DefaultValue = *field.DefaultValue
			}

			if field.ConditionalLogic != nil {
				logicJSON, err := json.Marshal(field.ConditionalLogic)
				if err != nil {
					return err
				}
				fieldParams.ConditionalLogic = logicJSON
			} else {
				fieldParams.ConditionalLogic = json.RawMessage("{}")
			}

			if field.FileConfig != nil {
				configJSON, err := json.Marshal(field.FileConfig)
				if err != nil {
					return err
				}
				fieldParams.FileConfig = configJSON
			} else {
				fieldParams.FileConfig = json.RawMessage("{}")
			}

			_, err = q.CreateFormField(ctx, fieldParams)
			if err != nil {
				return err
			}
		}

		// Create persistence config if provided
		if input.PersistenceConfig != nil {
			configParams := CreatePersistenceConfigParams{
				ID:               uuid.New(),
				FormDefinitionID: form.ID,
				PersistenceMode:  input.PersistenceConfig.PersistenceMode,
			}

			targetJSON, err := json.Marshal(input.PersistenceConfig.TargetConfigs)
			if err != nil {
				return err
			}
			configParams.TargetConfigs = targetJSON

			mappingsJSON, err := json.Marshal(input.PersistenceConfig.FieldMappings)
			if err != nil {
				return err
			}
			configParams.FieldMappings = mappingsJSON

			if input.PersistenceConfig.TransformationRules != nil {
				rulesJSON, err := json.Marshal(input.PersistenceConfig.TransformationRules)
				if err != nil {
					return err
				}
				configParams.TransformationRules = rulesJSON
			} else {
				configParams.TransformationRules = json.RawMessage("{}")
			}

			if input.PersistenceConfig.ValidationHooks != nil {
				hooksJSON, err := json.Marshal(input.PersistenceConfig.ValidationHooks)
				if err != nil {
					return err
				}
				configParams.ValidationHooks = hooksJSON
			} else {
				configParams.ValidationHooks = json.RawMessage("{}")
			}

			_, err = q.CreatePersistenceConfig(ctx, configParams)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &form, nil
}

// ProcessFormSubmissionTx processes a form submission with data persistence
func (store *SQLStore) ProcessFormSubmissionTx(ctx context.Context, input *FormSubmissionInput) (*FormSubmission, error) {
	var submission FormSubmission

	err := store.execTx(ctx, func(q *Queries) error {
		// Get persistence config
		config, err := q.GetPersistenceConfig(ctx, input.FormDefinitionID)
		if err != nil {
			return fmt.Errorf("failed to get persistence config: %w", err)
		}

		// Parse config
		var persistenceMode string
		var fieldMappings map[string]FieldMapping
		var targetConfigs []TargetConfig

		persistenceMode = config.PersistenceMode

		if err := json.Unmarshal(config.FieldMappings, &fieldMappings); err != nil {
			return fmt.Errorf("failed to parse field mappings: %w", err)
		}

		if err := json.Unmarshal(config.TargetConfigs, &targetConfigs); err != nil {
			return fmt.Errorf("failed to parse target configs: %w", err)
		}

		// Process based on persistence mode
		switch persistenceMode {
		case "direct":
			err = store.persistDirectMode(ctx, q, targetConfigs[0].TableName, fieldMappings, input.Data)
			if err != nil {
				return err
			}

		case "multi_table":
			err = store.persistMultiTableMode(ctx, q, fieldMappings, input.Data)
			if err != nil {
				return err
			}

		case "json":
			// Data will be stored in submission record
		}

		// Create submission record
		dataJSON, err := json.Marshal(input.Data)
		if err != nil {
			return err
		}

		metadataJSON, err := json.Marshal(input.Metadata)
		if err != nil {
			return err
		}

		submissionParams := CreateFormSubmissionParams{
			ID:               uuid.New(),
			FormDefinitionID: input.FormDefinitionID,
			UserID:           input.UserID,
			SubmissionData:   dataJSON,
			Status:           input.Status,
			Metadata:         metadataJSON,
		}

		createdSubmission, err := q.CreateFormSubmission(ctx, submissionParams)
		if err != nil {
			return err
		}
		submission = createdSubmission

		// Save files
		for _, file := range input.Files {
			fileParams := CreateFormSubmissionFileParams{
				ID:               uuid.New(),
				FormSubmissionID: submission.ID,
				FieldName:        file.FieldName,
				FileName:         file.FileName,
				FilePath:         file.FilePath,
				FileSize:         file.FileSize,
				MimeType:         file.MimeType,
				Bucket:           file.Bucket,
				StorageProvider:  file.StorageProvider,
			}

			_, err := q.CreateFormSubmissionFile(ctx, fileParams)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &submission, nil
}

// persistDirectMode saves to a single target table
func (store *SQLStore) persistDirectMode(ctx context.Context, q *Queries, tableName string, mappings map[string]FieldMapping, data map[string]interface{}) error {
	// Build dynamic SQL based on table name and mappings
	columns := []string{}
	values := []interface{}{}
	placeholders := []string{}

	i := 1
	for formField, value := range data {
		if mapping, ok := mappings[formField]; ok && mapping.TableName == tableName {
			columns = append(columns, mapping.ColumnName)
			values = append(values, value)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			i++
		}
	}

	if len(columns) == 0 {
		return nil
	}

	// Execute dynamic insert
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := q.db.ExecContext(ctx, query, values...)
	return err
}

// persistMultiTableMode saves to multiple tables
func (store *SQLStore) persistMultiTableMode(ctx context.Context, q *Queries, mappings map[string]FieldMapping, data map[string]interface{}) error {
	// Group by table
	tableData := make(map[string]map[string]interface{})

	for formField, value := range data {
		if mapping, ok := mappings[formField]; ok {
			if tableData[mapping.TableName] == nil {
				tableData[mapping.TableName] = make(map[string]interface{})
			}
			tableData[mapping.TableName][mapping.ColumnName] = value
		}
	}

	// Insert into each table
	for tableName, tableValues := range tableData {
		columns := []string{}
		values := []interface{}{}
		placeholders := []string{}

		i := 1
		for column, value := range tableValues {
			columns = append(columns, column)
			values = append(values, value)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			i++
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)

		_, err := q.db.ExecContext(ctx, query, values...)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateFormSubmissionTx updates a form submission with data persistence
func (store *SQLStore) UpdateFormSubmissionTx(ctx context.Context, input *FormSubmissionUpdateInput) (*FormSubmission, error) {
	var submission FormSubmission

	err := store.execTx(ctx, func(q *Queries) error {
		// Get persistence config

		// Parse config
		var fieldMappings map[string]FieldMapping
		var targetConfigs []TargetConfig
		config, err := q.GetPersistenceConfig(ctx, input.FormDefinitionID)
		if err != nil {
			// If no persistence config, just update submission
			// This is not an error - some forms only store in JSON
			goto updateSubmission
		}

		if err := json.Unmarshal(config.FieldMappings, &fieldMappings); err != nil {
			return fmt.Errorf("failed to parse field mappings: %w", err)
		}

		if err := json.Unmarshal(config.TargetConfigs, &targetConfigs); err != nil {
			return fmt.Errorf("failed to parse target configs: %w", err)
		}

		// Process based on persistence mode if status is changing to submitted
		if input.Status == "submitted" {
			switch config.PersistenceMode {
			case "direct":
				err = store.persistDirectMode(ctx, q, targetConfigs[0].TableName, fieldMappings, input.Data)
				if err != nil {
					return fmt.Errorf("failed to persist direct mode: %w", err)
				}

			case "multi_table":
				err = store.persistMultiTableMode(ctx, q, fieldMappings, input.Data)
				if err != nil {
					return fmt.Errorf("failed to persist multi-table mode: %w", err)
				}

			case "json":
				// Data will be stored in submission record
			}
		}

	updateSubmission:
		// Prepare final data
		var finalData map[string]interface{}
		if input.IsPartialUpdate && input.ExistingData != nil {
			// Merge with existing data
			if err := json.Unmarshal(input.ExistingData, &finalData); err != nil {
				finalData = make(map[string]interface{})
			}

			// Merge new data
			for key, value := range input.Data {
				finalData[key] = value
			}
		} else {
			finalData = input.Data
		}

		// Update submission record
		dataJSON, err := json.Marshal(finalData)
		if err != nil {
			return err
		}

		metadataJSON, err := json.Marshal(input.Metadata)
		if err != nil {
			return err
		}

		// Get existing submission to preserve approval fields
		existingSubmission, err := q.GetFormSubmission(ctx, input.SubmissionID)
		if err != nil {
			return err
		}

		submissionParams := UpdateFormSubmissionParams{
			ID:             input.SubmissionID,
			SubmissionData: dataJSON,
			Status:         input.Status,
			Metadata:       metadataJSON,
			ApprovalStatus: existingSubmission.ApprovalStatus,
			ApprovalNotes:  existingSubmission.ApprovalNotes,
			ApprovedBy:     existingSubmission.ApprovedBy,
			ApprovedAt:     existingSubmission.ApprovedAt,
		}

		updatedSubmission, err := q.UpdateFormSubmission(ctx, submissionParams)
		if err != nil {
			return err
		}
		submission = updatedSubmission

		// Save new files
		for _, file := range input.Files {
			fileParams := CreateFormSubmissionFileParams{
				ID:               uuid.New(),
				FormSubmissionID: submission.ID,
				FieldName:        file.FieldName,
				FileName:         file.FileName,
				FilePath:         file.FilePath,
				FileSize:         file.FileSize,
				MimeType:         file.MimeType,
				Bucket:           file.Bucket,
				StorageProvider:  file.StorageProvider,
			}

			_, err := q.CreateFormSubmissionFile(ctx, fileParams)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &submission, nil
}

func (store *SQLStore) SaveStepProgressTx(ctx context.Context, input *SaveStepProgressInput) error {
	return store.execTx(ctx, func(q *Queries) error {
		// Check if progress exists
		existing, err := q.GetStepProgress(ctx, GetStepProgressParams{
			FormSubmissionID: NewNullUUID(input.SubmissionID),
			FormStepID:       NewNullUUID(input.StepID),
		})

		dataJSON, err := json.Marshal(input.Data)
		if err != nil {
			return err
		}

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if existing.ID != uuid.Nil {
			// Update existing
			_, err = q.UpdateStepProgress(ctx, UpdateStepProgressParams{
				ID:     existing.ID,
				Status: NewNullString(input.Status),
				Data:   NullRawMessage(dataJSON),
			})
		} else {
			// Create new
			_, err = q.CreateStepProgress(ctx, CreateStepProgressParams{
				ID:               uuid.New(),
				FormSubmissionID: NewNullUUID(input.SubmissionID),
				FormStepID:       NewNullUUID(input.StepID),
				StepNumber:       input.StepNumber,
				Status:           NewNullString(input.Status),
				Data:             NullRawMessage(dataJSON),
			})
		}

		if err != nil {
			return err
		}

		// Also update main submission with partial data
		submission, err := q.GetFormSubmission(ctx, input.SubmissionID)
		if err != nil {
			return err
		}

		// Merge step data into submission data
		var existingData map[string]interface{}
		if submission.SubmissionData != nil {
			json.Unmarshal(submission.SubmissionData, &existingData)
		} else {
			existingData = make(map[string]interface{})
		}

		// Merge new data
		for key, value := range input.Data {
			existingData[key] = value
		}

		mergedJSON, err := json.Marshal(existingData)
		if err != nil {
			return err
		}

		// Update submission with merged data
		_, err = q.UpdateFormSubmission(ctx, UpdateFormSubmissionParams{
			ID:             input.SubmissionID,
			SubmissionData: mergedJSON,
			Status:         submission.Status, // Keep current status
			ApprovalStatus: submission.ApprovalStatus,
			ApprovalNotes:  submission.ApprovalNotes,
			ApprovedBy:     submission.ApprovedBy,
			ApprovedAt:     submission.ApprovedAt,
			Metadata:       submission.Metadata,
		})

		return err
	})
}
