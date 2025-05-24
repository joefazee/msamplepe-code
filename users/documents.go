package users

import (
	"fmt"
	"net/http"

	"github.com/timchuks/monieverse/core/controllers/shared"
	"github.com/timchuks/monieverse/internal/notifier"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/validator"
)

func (c *usersController) UploadIdentityDocument(ctx *gin.Context) {

	srv := c.srv

	authUser := srv.ContextGetUser(ctx)

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	req := domain.UploadIdentityDocumentRequest{
		ContentType:  header.Header.Get("Content-Type"),
		Size:         header.Size,
		FileName:     header.Filename,
		DocumentType: ctx.Request.FormValue("document_type"),
		Bucket:       srv.Config.FileBucket,
	}

	v := validator.New()
	if !req.Validate(v, srv.Config) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	dbRes, err := srv.Store.CreateIdentityDocumentTx(ctx, shared.UploadIdentityDocumentToSource(srv, authUser, file, req))

	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	srv.SendNotification(ctx,
		notifier.NewEmailRecipient(srv.Config.AdminEmail),
		"KYC Document Upload: "+authUser.AccountName(),
		fmt.Sprintf("The user %s with email %s just uploaded a KYC document", authUser.AccountName(), authUser.Email), nil,
	)

	srv.SuccessJSONResponse(ctx, http.StatusOK, "file uploaded successfully",
		domain.UploadIdentityDocumentResponse{
			ID: dbRes.ID.String(),
		})

}
