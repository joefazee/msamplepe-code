package forms

import (
	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/core/server"
	"github.com/timchuks/monieverse/internal/forms/handlers"
	"github.com/timchuks/monieverse/internal/forms/service"
)

func RegisterFormRoutes(r *gin.RouterGroup, srv *server.Server) {
	formService := service.NewFormService(
		srv.Store,
		srv.Uploader,
		nil,
		srv.TaskDistributor,
		&srv.Config,
		service.NewNullVirusScanner(),
		service.CreateContentValidator(),
		srv.Logger,
	)

	handler := handlers.NewFormHandler(srv, formService)

	// User routes
	userRoutes := r.Group("/forms")
	userRoutes.Use(srv.AuthenticatedUseRequired())
	userRoutes.Use(srv.ActivatedUserRequired())
	//srv.Idempotency(ratelimiter.OperationTypeFormSubmit, nil),
	userRoutes.GET("/", handler.GetUserForm) // ?type=kyc|kyb|contact etc

	// CREATE: Start new form submission as draft
	// POST /forms/{form_id}/draft
	// Content-Type: multipart/form-data
	// Body: business_name=Acme&_status=draft&_partial=true
	userRoutes.POST("/:id/draft", handler.CreateDraftSubmission)

	// CREATE: Submit complete form directly (bypasses draft)
	// POST /forms/{form_id}/submit
	// Content-Type: multipart/form-data
	// Body: all_required_fields + files + _status=submitted
	userRoutes.POST("/:id/submit",
		handler.SubmitForm)

	userRoutes.GET("/submissions", handler.GetFormSubmissions)

	// READ: Get submission for editing (includes form definition + existing data)
	userRoutes.GET("/submissions/:id", handler.GetFormSubmission)

	// Add to RegisterFormRoutes
	userRoutes.GET("/submissions/:id/edit", handler.GetFormSubmissionForEdit)

	// UPDATE: Update entire submission
	// PUT /submissions/{id}
	// Content-Type: multipart/form-data
	// Body: updated_fields + files + _status=draft|submitted + _partial=true|false
	userRoutes.PUT("/submissions/:id",
		handler.UpdateFormSubmission)

	// DELETE: Remove uploaded files
	userRoutes.DELETE("/submissions/:id/files/:fileId", handler.DeleteFormSubmissionFile)

	userRoutes.GET("/progress", handler.GetFormWithProgress) // ?type=kyb

	// UPDATE: Save progress for specific step
	// PUT /submissions/{id}/steps/{step}
	// Content-Type: multipart/form-data
	// Body: step_fields + files + _status=in_progress|completed
	userRoutes.PUT("/submissions/:id/steps/:step", handler.SaveStepProgress)

	// READ: Get data for specific step
	userRoutes.GET("/submissions/:id/steps/:step", handler.GetStepData)

	// CREATE: Complete all steps and finalize submission
	// POST /submissions/{id}/complete
	// Content-Type: multipart/form-data
	// Body: _status=submitted (optional additional data)
	userRoutes.POST("/submissions/:id/complete", handler.CompleteForm)

	// Admin routes
	adminRoutes := r.Group("/admin/forms")
	adminRoutes.Use(srv.AuthenticatedUseRequired())
	adminRoutes.Use(srv.RequirePermission("admin.forms"))

	// Form Definition Management
	adminRoutes.POST("/", handler.CreateFormDefinition)
	adminRoutes.PUT("/:id", handler.UpdateFormDefinition)
	adminRoutes.GET("/:id", handler.GetFormDefinition)
	adminRoutes.GET("/", handler.ListFormDefinitions)

	// Form Assignment Management
	adminRoutes.POST("/:id/assignments", handler.CreateFormAssignment)
	adminRoutes.GET("/:id/assignments", handler.GetFormAssignments)

	// Persistence Configuration
	adminRoutes.POST("/:id/persistence", handler.CreatePersistenceConfig)
	adminRoutes.PUT("/:id/persistence", handler.UpdatePersistenceConfig)

	// Submission Management & Approval
	adminRoutes.POST("/submissions/:id/approve", handler.ApproveSubmission)
	adminRoutes.POST("/submissions/:id/reject", handler.RejectSubmission)
}
