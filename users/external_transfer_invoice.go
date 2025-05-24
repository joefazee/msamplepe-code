package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/timchuks/monieverse/core/controllers/shared"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/notifier"
	"github.com/timchuks/monieverse/internal/validator"
)

const (
	invoicesPath = "transfer/invoice"
)

func (c *usersController) UploadTransferInvoice(ctx *gin.Context) {
	srv := c.srv
	authUser := srv.ContextGetUser(ctx)

	transferID, err := uuid.Parse(ctx.Request.FormValue("transfer_id"))
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid transfer id"))
		return
	}

	transfer, err := c.findUserTransfer(ctx, transferID, authUser)
	if err != nil {
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	oldDoc, err := srv.Store.GetUserDocument(ctx, db.GetUserDocumentParams{
		UserID:  authUser.ID,
		Model:   db.DocumentModelTransferInvoice,
		ModelID: transfer.ID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		srv.Logger.Error(err, map[string]interface{}{
			"user_id": authUser.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to get user document"))
		return
	}

	req := domain.UploadDocumentRequest{
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		FileName:    header.Filename,
		Bucket:      srv.Config.FileBucket,
		Model:       db.DocumentModelTransferInvoice,
		ModelID:     transfer.ID,
	}

	v := validator.New()
	if !req.Validate(v, srv.Config) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	dbRes, err := srv.Store.CreateDocumentTx(ctx, shared.UploadDocumentToSource(srv, authUser, file, req, fmt.Sprintf("%s/%s", invoicesPath, transfer.ID.String())))

	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	if err = shared.CleanUpOldDocument(ctx, srv, oldDoc); err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"doc": oldDoc,
		})
	}

	if _, err := srv.Store.UpdateTransactionInvoiceStatus(ctx, transferID); err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"error": err,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	srv.SendNotification(ctx,
		notifier.NewEmailRecipient(srv.Config.AdminEmail),
		"Transfer Invoice: "+authUser.AccountName(),
		fmt.Sprintf("The user %s with email %s just uploaded a transfer invoice", authUser.AccountName(), authUser.Email), nil,
	)

	_ = srv.Store.CreateTransactionHistory(ctx, db.CreateTransactionHistoryParams{
		TransactionID: transfer.ID,
		UserID:        transfer.UserID,
		Reason:        "invoid uploaded",
		Amount:        transfer.Amount,
		OldStatus:     transfer.Status,
		NewStatus:     transfer.Status,
	})

	srv.SuccessJSONResponse(ctx, http.StatusOK, "file uploaded successfully",
		domain.UploadIdentityDocumentResponse{
			ID: dbRes.ID.String(),
		})

}

func (c *usersController) findUserTransfer(ctx *gin.Context, transferID uuid.UUID, authUser *db.User) (*db.Transaction, error) {
	srv := c.srv

	transfer, err := srv.Store.GetUserTransactionByManyStatus(ctx, db.GetUserTransactionByManyStatusParams{
		ID:      transferID,
		UserID:  authUser.ID,
		Column3: []string{db.TransactionStatusPending, db.TransactionStatusProcessing},
	})

	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"transfer_id": transferID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("transfer not found"))
		return nil, err
	}

	return &transfer, nil
}
