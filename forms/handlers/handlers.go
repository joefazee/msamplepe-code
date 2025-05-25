package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/timchuks/monieverse/core/server"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/forms/service"
)

type FormHandler struct {
	srv         *server.Server
	formService *service.FormService
}

// StandardFormData represents the standardized form data structure
type StandardFormData struct {
	Fields map[string]interface{}             `json:"fields"`
	Files  map[string][]*multipart.FileHeader `json:"-"`
	Meta   FormMeta                           `json:"meta"`
}

// FormMeta contains metadata about the form submission
type FormMeta struct {
	Status         string                 `json:"status"`          // draft, in_progress, completed, submitted
	IsPartial      bool                   `json:"is_partial"`      // Whether this is a partial update
	StepNumber     *int32                 `json:"step_number"`     // For step-specific operations
	SkipValidation bool                   `json:"skip_validation"` // Skip validation (admin only)
	Metadata       map[string]interface{} `json:"metadata"`        // Additional metadata
}

func NewFormHandler(srv *server.Server, formService *service.FormService) *FormHandler {
	return &FormHandler{
		srv:         srv,
		formService: formService,
	}
}

func (h *FormHandler) CreateFormDefinition(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	var req CreateFormRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Build input
	input := &db.FormDefinitionInput{
		Name:                      req.Name,
		Slug:                      req.Slug,
		Description:               req.Description,
		FormType:                  req.FormType,
		IsMultiStep:               req.IsMultiStep,
		RequiresApproval:          req.RequiresApproval,
		IsEditableAfterSubmission: req.IsEditableAfterSubmission,
		CreatedBy:                 user.ID,
	}

	// Convert ApprovalWorkflow
	if req.ApprovalWorkflow != nil {
		approvalWorkflow := &db.ApprovalWorkflowInput{
			States:      make([]map[string]interface{}, len(req.ApprovalWorkflow.States)),
			Transitions: make([]map[string]interface{}, len(req.ApprovalWorkflow.Transitions)),
		}

		// Convert states
		for i, state := range req.ApprovalWorkflow.States {
			approvalWorkflow.States[i] = map[string]interface{}{
				"name":        state.Name,
				"label":       state.Label,
				"is_final":    state.IsFinal,
				"permissions": state.Permissions,
			}
		}

		// Convert transitions
		for i, transition := range req.ApprovalWorkflow.Transitions {
			approvalWorkflow.Transitions[i] = map[string]interface{}{
				"from":        transition.From,
				"to":          transition.To,
				"label":       transition.Label,
				"permissions": transition.Permissions,
			}
		}

		input.ApprovalWorkflow = approvalWorkflow
	}

	// Add steps
	for _, step := range req.Steps {
		input.Steps = append(input.Steps, db.StepInput{
			StepNumber:  step.StepNumber,
			Name:        step.Name,
			Description: step.Description,
			IsOptional:  step.IsOptional,
		})
	}

	// Add fields
	for _, field := range req.Fields {
		fieldInput := db.FieldInput{
			FieldName:    field.FieldName,
			FieldType:    field.FieldType,
			Label:        field.Label,
			DisplayOrder: field.DisplayOrder,
			IsRequired:   field.IsRequired,
			IsReadonly:   field.IsReadonly,
		}

		if field.StepNumber != nil {
			fieldInput.StepNumber = *field.StepNumber
		}

		if field.Placeholder != nil {
			fieldInput.Placeholder = field.Placeholder
		}

		if field.HelpText != nil {
			fieldInput.HelpText = field.HelpText
		}

		if field.ValidationRules != nil {
			fieldInput.ValidationRules = field.ValidationRules
		}

		// Convert Options to map
		if field.Options != nil {
			optionsMap := map[string]interface{}{
				"type": field.Options.Type,
			}

			if len(field.Options.Static) > 0 {
				staticOptions := make([]map[string]interface{}, len(field.Options.Static))
				for i, opt := range field.Options.Static {
					staticOptions[i] = map[string]interface{}{
						"value": opt.Value,
						"label": opt.Label,
					}
				}
				optionsMap["static"] = staticOptions
			}

			if field.Options.Dynamic != nil {
				dynamicMap := map[string]interface{}{
					"source_name": field.Options.Dynamic.SourceName,
				}
				if field.Options.Dynamic.FilterParams != nil {
					dynamicMap["filter_params"] = field.Options.Dynamic.FilterParams
				}
				optionsMap["dynamic"] = dynamicMap
			}

			fieldInput.Options = optionsMap
		}

		if field.DefaultValue != nil {
			fieldInput.DefaultValue = field.DefaultValue
		}

		// Convert ConditionalLogic to map
		if field.ConditionalLogic != nil {
			conditions := make([]map[string]interface{}, len(field.ConditionalLogic.Conditions))
			for i, cond := range field.ConditionalLogic.Conditions {
				conditions[i] = map[string]interface{}{
					"field":    cond.Field,
					"operator": cond.Operator,
					"value":    cond.Value,
				}
			}

			fieldInput.ConditionalLogic = map[string]interface{}{
				"action":     field.ConditionalLogic.Action,
				"conditions": conditions,
				"logic":      field.ConditionalLogic.Logic,
			}
		}

		// Convert FileConfig to map
		if field.FileConfig != nil {
			fileConfigMap := map[string]interface{}{
				"max_size":      field.FileConfig.MaxSize,
				"allowed_types": field.FileConfig.AllowedTypes,
			}
			if field.FileConfig.MaxFiles > 0 {
				fileConfigMap["max_files"] = field.FileConfig.MaxFiles
			}
			fieldInput.FileConfig = fileConfigMap
		}

		input.Fields = append(input.Fields, fieldInput)
	}

	// Convert PersistenceConfig
	if req.PersistenceConfig != nil {
		persistenceConfig := &db.PersistenceConfigInput{
			PersistenceMode: req.PersistenceConfig.PersistenceMode,
		}

		// Convert TargetConfigs
		targetConfigs := make([]map[string]interface{}, len(req.PersistenceConfig.TargetConfigs))
		for i, target := range req.PersistenceConfig.TargetConfigs {
			targetConfigs[i] = map[string]interface{}{
				"table_name": target.TableName,
				"priority":   target.Priority,
			}
			if target.Conditions != nil {
				targetConfigs[i]["conditions"] = target.Conditions
			}
		}
		persistenceConfig.TargetConfigs = targetConfigs

		// Convert FieldMappings
		fieldMappings := make(map[string]interface{})
		for key, mapping := range req.PersistenceConfig.FieldMappings {
			mappingMap := map[string]interface{}{
				"form_field":   mapping.FormField,
				"table_name":   mapping.TableName,
				"column_name":  mapping.ColumnName,
				"data_type":    mapping.DataType,
				"is_encrypted": mapping.IsEncrypted,
			}
			if mapping.Transform != nil {
				mappingMap["transform"] = *mapping.Transform
			}
			if mapping.MetaKey != "" {
				mappingMap["meta_key"] = mapping.MetaKey
			}
			fieldMappings[key] = mappingMap
		}
		persistenceConfig.FieldMappings = fieldMappings

		if req.PersistenceConfig.TransformationRules != nil {
			persistenceConfig.TransformationRules = req.PersistenceConfig.TransformationRules
		}

		// Convert ValidationHooks
		if req.PersistenceConfig.ValidationHooks != nil {
			validationHooks := make([]map[string]interface{}, len(req.PersistenceConfig.ValidationHooks))
			for i, hook := range req.PersistenceConfig.ValidationHooks {
				validationHooks[i] = map[string]interface{}{
					"type":   hook.Type,
					"config": hook.Config,
				}
			}
			persistenceConfig.ValidationHooks = validationHooks
		}

		input.PersistenceConfig = persistenceConfig
	}

	// Create form
	form, err := h.srv.Store.CreateFormDefinitionTx(ctx, input)
	if err != nil {
		h.srv.Logger.Error(err, map[string]interface{}{
			"request": req,
		})
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusCreated, "Form created successfully", form)
}

// UpdateFormDefinition updates a form
func (h *FormHandler) UpdateFormDefinition(ctx *gin.Context) {
	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	var req UpdateFormRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Update form
	params := db.UpdateFormDefinitionParams{
		ID:                        formID,
		Name:                      req.Name,
		Description:               req.Description,
		IsActive:                  req.IsActive,
		IsEditableAfterSubmission: req.IsEditableAfterSubmission,
	}

	if req.ApprovalWorkflow != nil {
		workflowJSON, _ := json.Marshal(req.ApprovalWorkflow)
		params.ApprovalWorkflow = workflowJSON
	}

	form, err := h.srv.Store.UpdateFormDefinition(ctx, params)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form updated successfully", form)
}

// GetFormDefinition retrieves a form
func (h *FormHandler) GetFormDefinition(ctx *gin.Context) {
	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	form, err := h.srv.Store.GetFormDefinition(ctx, formID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("form not found"))
			return
		}
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// Get steps
	steps, _ := h.srv.Store.GetFormSteps(ctx, formID)

	// Get fields
	fields, _ := h.srv.Store.GetFormFields(ctx, formID)

	response := map[string]interface{}{
		"form":   form,
		"steps":  steps,
		"fields": fields,
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form retrieved successfully", response)
}

// ListFormDefinitions lists all forms
func (h *FormHandler) ListFormDefinitions(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	formType := ctx.Query("form_type")
	isActiveStr := ctx.Query("is_active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := db.ListFormDefinitionsParams{
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	}

	if formType != "" {
		params.Column1 = formType
	}

	if isActiveStr != "" {
		isActive := isActiveStr == "true"
		params.Column2 = isActive
	}

	forms, err := h.srv.Store.ListFormDefinitions(ctx, params)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Forms retrieved successfully", forms)
}

// CreateFormAssignment assigns a form
func (h *FormHandler) CreateFormAssignment(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	var req CreateAssignmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	params := db.CreateFormAssignmentParams{
		ID:               uuid.New(),
		FormDefinitionID: formID,
		AssignmentType:   req.AssignmentType,
		AssignmentValue:  req.AssignmentValue,
		Priority:         int32(req.Priority),
		CreatedBy:        db.NewNullUUID(user.ID),
	}

	if req.Conditions != nil {
		conditionsJSON, _ := json.Marshal(req.Conditions)
		params.Conditions.RawMessage = conditionsJSON
	}

	if req.ValidFrom != nil {
		params.ValidFrom = db.NewNullTime(*req.ValidFrom)
	}

	if req.ValidUntil != nil {
		params.ValidUntil = db.NewNullTime(*req.ValidUntil)
	}

	assignment, err := h.srv.Store.CreateFormAssignment(ctx, params)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusCreated, "Assignment created successfully", assignment)
}

// CreatePersistenceConfig creates persistence config
func (h *FormHandler) CreatePersistenceConfig(ctx *gin.Context) {
	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	var req PersistenceConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	params := db.CreatePersistenceConfigParams{
		ID:               uuid.New(),
		FormDefinitionID: formID,
		PersistenceMode:  req.PersistenceMode,
	}

	targetJSON, _ := json.Marshal(req.TargetConfigs)
	params.TargetConfigs = targetJSON

	mappingsJSON, _ := json.Marshal(req.FieldMappings)
	params.FieldMappings = mappingsJSON

	if req.TransformationRules != nil {
		rulesJSON, _ := json.Marshal(req.TransformationRules)
		params.TransformationRules = rulesJSON
	}

	if req.ValidationHooks != nil {
		hooksJSON, _ := json.Marshal(req.ValidationHooks)
		params.ValidationHooks = hooksJSON
	}

	config, err := h.srv.Store.CreatePersistenceConfig(ctx, params)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusCreated, "Persistence config created successfully", config)
}

// User Endpoints

// GetUserForm retrieves the appropriate form for a user
func (h *FormHandler) GetUserForm(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)
	formType := ctx.Query("type")

	if formType == "" {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("form type is required"))
		return
	}

	form, err := h.formService.GetFormForUser(ctx, user.ID, formType)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form retrieved successfully", form)
}

// SubmitForm handles form submission
func (h *FormHandler) SubmitForm(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Parse multipart form
	if err := ctx.Request.ParseMultipartForm(h.srv.Config.MaxBodyPayloadSize); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Extract form data
	data := make(map[string]interface{})
	for key, values := range ctx.Request.MultipartForm.Value {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	// Parse numbers and booleans
	for key, value := range data {
		if str, ok := value.(string); ok {
			// Try to parse as number
			if num, err := strconv.ParseFloat(str, 64); err == nil {
				data[key] = num
			} else if str == "true" || str == "false" {
				data[key] = str == "true"
			}
		}
	}

	input := service.SubmitFormInput{
		FormID: formID,
		UserID: user.ID,
		Data:   data,
		Files:  ctx.Request.MultipartForm.File,
		Metadata: map[string]interface{}{
			"ip_address": ctx.ClientIP(),
			"user_agent": ctx.Request.UserAgent(),
		},
	}

	submission, err := h.formService.SubmitForm(ctx, input)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form submitted successfully", submission)
}

// GetFormSubmission retrieves a specific submission
func (h *FormHandler) GetFormSubmission(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	submission, err := h.srv.Store.GetFormSubmission(ctx, submissionID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("submission not found"))
			return
		}
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// Check ownership
	if submission.UserID != user.ID && !h.srv.Store.HasPermission(ctx, *user, "admin.forms") {
		h.srv.ErrorJSONResponse(ctx, http.StatusForbidden, fmt.Errorf("access denied"))
		return
	}

	// Get files
	files, _ := h.srv.Store.GetFormSubmissionFiles(ctx, submissionID)

	response := map[string]interface{}{
		"submission": submission,
		"files":      files,
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Submission retrieved successfully", response)
}

// GetFormSubmissions lists user's submissions
func (h *FormHandler) GetFormSubmissions(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	status := ctx.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := db.ListFormSubmissionsParams{
		Column1: user.ID,
		Limit:   int32(pageSize),
		Offset:  int32((page - 1) * pageSize),
	}

	if status != "" {
		params.Column3 = status
	}

	submissions, err := h.srv.Store.ListFormSubmissions(ctx, params)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Submissions retrieved successfully", submissions)
}

// ApproveSubmission approves a submission
func (h *FormHandler) ApproveSubmission(ctx *gin.Context) {
	approver := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	var req ApprovalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.formService.ApproveSubmission(ctx, submissionID, approver.ID, req.Notes); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Submission approved successfully", nil)
}

// RejectSubmission rejects a submission
func (h *FormHandler) RejectSubmission(ctx *gin.Context) {
	approver := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	var req ApprovalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	if req.Reason == "" {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("reason is required"))
		return
	}

	if err := h.formService.RejectSubmission(ctx, submissionID, approver.ID, req.Reason); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Submission rejected successfully", nil)
}

// Add these methods to FormHandler in forms/handlers package

// GetFormAssignments retrieves form assignments
func (h *FormHandler) GetFormAssignments(ctx *gin.Context) {
	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Get all assignments for this form
	assignments, err := h.srv.Store.GetFormAssignments(ctx, db.GetFormAssignmentsParams{
		AssignmentValue:   "", // Not filtering by user
		AssignmentValue_2: "", // Not filtering by type
		AssignmentValue_3: "", // Not filtering by country
		AssignmentValue_4: "", // Not filtering by state
	})

	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// Filter assignments for this form
	var formAssignments []db.GetFormAssignmentsRow
	for _, assignment := range assignments {
		if assignment.FormDefinitionID == formID {
			formAssignments = append(formAssignments, assignment)
		}
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Assignments retrieved successfully", formAssignments)
}

// UpdatePersistenceConfig updates persistence configuration
func (h *FormHandler) UpdatePersistenceConfig(ctx *gin.Context) {
	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	var req PersistenceConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Check if config exists
	_, err = h.srv.Store.GetPersistenceConfig(ctx, formID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("persistence config not found"))
			return
		}
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// Build update params
	params := db.UpdatePersistenceConfigParams{
		FormDefinitionID: formID,
		PersistenceMode:  req.PersistenceMode,
	}

	targetJSON, _ := json.Marshal(req.TargetConfigs)
	params.TargetConfigs = targetJSON

	mappingsJSON, _ := json.Marshal(req.FieldMappings)
	params.FieldMappings = mappingsJSON

	if req.TransformationRules != nil {
		rulesJSON, _ := json.Marshal(req.TransformationRules)
		params.TransformationRules = rulesJSON
	}

	if req.ValidationHooks != nil {
		hooksJSON, _ := json.Marshal(req.ValidationHooks)
		params.ValidationHooks = hooksJSON
	}

	config, err := h.srv.Store.UpdatePersistenceConfig(ctx, params)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Persistence config updated successfully", config)
}

// GetFormSubmissionForEdit retrieves a form submission for editing
func (h *FormHandler) GetFormSubmissionForEdit(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	formData, err := h.formService.GetFormForEdit(ctx, submissionID, user.ID)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form retrieved successfully", formData)
}

// UpdateFormSubmission updates an existing form submission
func (h *FormHandler) UpdateFormSubmission(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Parse multipart form
	if err = ctx.Request.ParseMultipartForm(h.srv.Config.MaxBodyPayloadSize); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Extract form data
	data := make(map[string]interface{})
	for key, values := range ctx.Request.MultipartForm.Value {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	// Parse numbers and booleans
	for key, value := range data {
		if str, ok := value.(string); ok {
			// Try to parse as number
			if num, err := strconv.ParseFloat(str, 64); err == nil {
				data[key] = num
			} else if str == "true" || str == "false" {
				data[key] = str == "true"
			}
		}
	}

	// Get status from form data or default to current
	status := "draft"
	if statusValue, ok := data["_status"]; ok {
		status = statusValue.(string)
		delete(data, "_status") // Remove from data
	}

	// Check if partial update
	isPartialUpdate := false
	if partialValue, ok := data["_partial"]; ok {
		isPartialUpdate = partialValue.(string) == "true"
		delete(data, "_partial") // Remove from data
	}

	input := service.UpdateSubmissionInput{
		SubmissionID:    submissionID,
		UserID:          user.ID,
		Data:            data,
		Files:           ctx.Request.MultipartForm.File,
		Status:          status,
		IsPartialUpdate: isPartialUpdate,
		Metadata: map[string]interface{}{
			"ip_address": ctx.ClientIP(),
			"user_agent": ctx.Request.UserAgent(),
			"updated_at": time.Now(),
		},
	}

	submission, err := h.formService.UpdateFormSubmission(ctx, input)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form updated successfully", submission)
}

// DeleteFormSubmissionFile deletes a file from a submission
func (h *FormHandler) DeleteFormSubmissionFile(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	fileID, err := uuid.Parse(ctx.Param("fileId"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Verify ownership
	submission, err := h.srv.Store.GetFormSubmission(ctx, submissionID)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("submission not found"))
		return
	}

	if submission.UserID != user.ID && !h.srv.Store.HasPermission(ctx, *user, "admin.forms") {
		h.srv.ErrorJSONResponse(ctx, http.StatusForbidden, fmt.Errorf("access denied"))
		return
	}

	// Get file
	files, err := h.srv.Store.GetFormSubmissionFiles(ctx, submissionID)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	var fileToDelete *db.FormSubmissionFile
	for _, file := range files {
		if file.ID == fileID {
			fileToDelete = &file
			break
		}
	}

	if fileToDelete == nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("file not found"))
		return
	}

	// Delete from storage
	if err := h.srv.Uploader.Delete(fileToDelete.Bucket, fileToDelete.FilePath); err != nil {
		h.srv.Logger.Error(err, map[string]interface{}{
			"file": fileToDelete,
		})
	}

	//TODO: Delete from database (I will add this later)
	// err = h.srv.Store.DeleteFormSubmissionFile(ctx, fileID)

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "File deleted successfully", nil)
}

// SaveStepProgress saves progress for a form step
func (h *FormHandler) SaveStepProgress(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	stepNumber, err := strconv.Atoi(ctx.Param("step"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Parse multipart form
	if err = ctx.Request.ParseMultipartForm(h.srv.Config.MaxBodyPayloadSize); err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Extract form data
	data := make(map[string]interface{})
	for key, values := range ctx.Request.MultipartForm.Value {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	// Handle file uploads for this step
	if len(ctx.Request.MultipartForm.File) > 0 {
		log.Printf("handle file upload")
		// Process files and update data with file references
		// Similar to regular form submission
	}

	status := ctx.DefaultQuery("status", "in_progress")

	input := service.SaveStepProgressInput{
		SubmissionID: submissionID,
		UserID:       user.ID,
		StepNumber:   int32(stepNumber),
		Status:       status,
		Data:         data,
	}

	result, err := h.formService.SaveStepProgress(ctx, input)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Step progress saved", result)
}

// GetFormWithProgress retrieves form with progress tracking
func (h *FormHandler) GetFormWithProgress(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)
	formType := ctx.Query("type")

	if formType == "" {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("form type is required"))
		return
	}

	form, err := h.formService.GetFormWithProgress(ctx, user.ID, formType)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form retrieved successfully", form)
}

// CompleteForm marks the form as complete after all steps
func (h *FormHandler) CompleteForm(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	submission, err := h.formService.CompleteForm(ctx, submissionID, user.ID)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form completed successfully", submission)
}

// GetStepData retrieves data for a specific step
func (h *FormHandler) GetStepData(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	stepNumber, err := strconv.Atoi(ctx.Param("step"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Verify ownership
	submission, err := h.srv.Store.GetFormSubmission(ctx, submissionID)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("submission not found"))
		return
	}

	if submission.UserID != user.ID {
		h.srv.ErrorJSONResponse(ctx, http.StatusForbidden, fmt.Errorf("access denied"))
		return
	}

	// Get step progress
	progress, err := h.srv.Store.GetStepProgressBySubmissionAndNumber(ctx,
		db.GetStepProgressBySubmissionAndNumberParams{
			FormSubmissionID: db.NewNullUUID(submissionID),
			StepNumber:       int32(stepNumber),
		})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.srv.SuccessJSONResponse(ctx, http.StatusOK, "No data for this step", map[string]interface{}{
				"status": "not_started",
				"data":   map[string]interface{}{},
			})
			return
		}
		h.srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	var stepData map[string]interface{}
	_ = json.Unmarshal(progress.Data.RawMessage, &stepData)

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Step data retrieved", map[string]interface{}{
		"status":       progress.Status,
		"data":         stepData,
		"completed_at": progress.CompletedAt,
	})
}

// parseStandardFormData extracts form data in standardized format
func (h *FormHandler) parseStandardFormData(ctx *gin.Context) (*StandardFormData, error) {
	// Parse multipart form
	err := ctx.Request.ParseMultipartForm(h.srv.Config.MaxBodyPayloadSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	data := &StandardFormData{
		Fields: make(map[string]interface{}),
		Files:  ctx.Request.MultipartForm.File,
		Meta: FormMeta{
			Status:    "draft", // Default status
			IsPartial: false,
			Metadata:  make(map[string]interface{}),
		},
	}

	// Extract regular form fields
	for key, values := range ctx.Request.MultipartForm.Value {
		// Handle meta fields with underscore prefix
		if strings.HasPrefix(key, "_") {
			h.parseMetaField(key, values, &data.Meta)
			continue
		}

		// Handle regular fields
		if len(values) == 1 {
			data.Fields[key] = h.parseFieldValue(values[0])
		} else if len(values) > 1 {
			// Handle multi-value fields (checkboxes, multi-select)
			parsedValues := make([]interface{}, len(values))
			for i, v := range values {
				parsedValues[i] = h.parseFieldValue(v)
			}
			data.Fields[key] = parsedValues
		}
	}

	// Add request metadata
	data.Meta.Metadata["ip_address"] = ctx.ClientIP()
	data.Meta.Metadata["user_agent"] = ctx.Request.UserAgent()
	data.Meta.Metadata["timestamp"] = time.Now()

	return data, nil
}

// parseMetaField handles meta fields like _status, _partial, etc.
func (h *FormHandler) parseMetaField(key string, values []string, meta *FormMeta) {
	if len(values) == 0 {
		return
	}

	value := values[0]

	switch key {
	case "_status":
		meta.Status = value
	case "_partial":
		meta.IsPartial = value == "true"
	case "_step":
		if stepNum, err := strconv.ParseInt(value, 10, 32); err == nil {
			step := int32(stepNum)
			meta.StepNumber = &step
		}
	case "_skip_validation":
		meta.SkipValidation = value == "true"
	default:
		// Store other meta fields in metadata
		metaKey := strings.TrimPrefix(key, "_")
		meta.Metadata[metaKey] = value
	}
}

// parseFieldValue attempts to parse string values to appropriate types
func (h *FormHandler) parseFieldValue(value string) interface{} {
	// Try to parse as number
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		// Check if it's actually an integer
		if num == float64(int64(num)) {
			return int64(num)
		}
		return num
	}

	// Try to parse as boolean
	if value == "true" || value == "false" {
		return value == "true"
	}

	// Try to parse as JSON (for complex objects)
	if (strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")) ||
		(strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]")) {
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			return jsonValue
		}
	}

	// Return as string if no parsing succeeded
	return value
}

// CreateDraftSubmission creates a new draft submission
func (h *FormHandler) CreateDraftSubmission(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid form ID"))
		return
	}

	// Parse form data
	formData, err := h.parseStandardFormData(ctx)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Force draft status for this endpoint
	formData.Meta.Status = "draft"

	// Create submission input
	input := service.CreateSubmissionInput{
		FormID:     formID,
		UserID:     user.ID,
		Data:       formData.Fields,
		Files:      formData.Files,
		Status:     formData.Meta.Status,
		IsPartial:  formData.Meta.IsPartial,
		Metadata:   formData.Meta.Metadata,
		StepNumber: formData.Meta.StepNumber,
	}

	submission, err := h.formService.CreateSubmission(ctx, input)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	response := map[string]interface{}{
		"submission": submission,
		"message":    "Draft submission created successfully",
		"status":     "draft",
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusCreated, "Draft created successfully", response)
}

// SubmitFormFinal submits a complete form (replaces old SubmitForm)
func (h *FormHandler) SubmitFormFinal(ctx *gin.Context) {
	user := h.srv.ContextGetUser(ctx)

	formID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid form ID"))
		return
	}

	// Parse form data
	formData, err := h.parseStandardFormData(ctx)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// Force submitted status for final submission
	formData.Meta.Status = "submitted"

	input := service.CreateSubmissionInput{
		FormID:     formID,
		UserID:     user.ID,
		Data:       formData.Fields,
		Files:      formData.Files,
		Status:     formData.Meta.Status,
		IsPartial:  false, // Final submissions are never partial
		Metadata:   formData.Meta.Metadata,
		StepNumber: formData.Meta.StepNumber,
	}

	submission, err := h.formService.CreateSubmission(ctx, input)
	if err != nil {
		h.srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	response := map[string]interface{}{
		"submission": submission,
		"message":    "Form submitted successfully",
		"status":     "submitted",
	}

	h.srv.SuccessJSONResponse(ctx, http.StatusOK, "Form submitted successfully", response)
}
