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
	)

	handler := handlers.NewFormHandler(srv, formService)

	// User routes
	userRoutes := r.Group("/forms")
	userRoutes.Use(srv.AuthenticatedUseRequired())
	userRoutes.Use(srv.ActivatedUserRequired())
	//srv.Idempotency(ratelimiter.OperationTypeFormSubmit, nil),
	userRoutes.GET("/", handler.GetUserForm) // ?type=kyc
	userRoutes.POST("/:id/submit",
		handler.SubmitForm)
	userRoutes.GET("/submissions", handler.GetFormSubmissions)
	userRoutes.GET("/submissions/:id", handler.GetFormSubmission)

	// Add to RegisterFormRoutes
	userRoutes.GET("/submissions/:id/edit", handler.GetFormSubmissionForEdit)
	userRoutes.PUT("/submissions/:id",
		handler.UpdateFormSubmission)
	userRoutes.DELETE("/submissions/:id/files/:fileId", handler.DeleteFormSubmissionFile)

	userRoutes.GET("/progress", handler.GetFormWithProgress) // ?type=kyb
	userRoutes.PUT("/submissions/:id/steps/:step", handler.SaveStepProgress)
	userRoutes.GET("/submissions/:id/steps/:step", handler.GetStepData)
	userRoutes.POST("/submissions/:id/complete", handler.CompleteForm)

	// Admin routes
	adminRoutes := r.Group("/admin/forms")
	adminRoutes.Use(srv.AuthenticatedUseRequired())
	adminRoutes.Use(srv.RequirePermission("admin.forms"))

	adminRoutes.POST("/", handler.CreateFormDefinition)
	adminRoutes.PUT("/:id", handler.UpdateFormDefinition)
	adminRoutes.GET("/:id", handler.GetFormDefinition)
	adminRoutes.GET("/", handler.ListFormDefinitions)

	adminRoutes.POST("/:id/assignments", handler.CreateFormAssignment)
	adminRoutes.GET("/:id/assignments", handler.GetFormAssignments)

	adminRoutes.POST("/:id/persistence", handler.CreatePersistenceConfig)
	adminRoutes.PUT("/:id/persistence", handler.UpdatePersistenceConfig)

	adminRoutes.POST("/submissions/:id/approve", handler.ApproveSubmission)
	adminRoutes.POST("/submissions/:id/reject", handler.RejectSubmission)
}
