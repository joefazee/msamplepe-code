package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/timchuks/monieverse/core/controllers/shared"
	"github.com/timchuks/monieverse/external/zylalabs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/timchuks/monieverse/internal/core"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/validator"
)

func (c *usersController) CreateRecipient(ctx *gin.Context) {

	srv := c.srv

	user := srv.ContextGetUser(ctx)

	scheme := ctx.Param("scheme")
	currencyCode := ctx.Param("currency")

	req, ok := shared.SupportedSchemeRequests[strings.ToUpper(scheme)]
	if !ok {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid scheme"))
		return
	}

	req.SetCurrency(currencyCode)
	req.SetScheme(scheme)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	v := validator.NewWithStore(ctx, srv.Store)
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	res, err := srv.Store.CreateRecipient(ctx, db.CreateRecipientParams{
		UserID:   user.ID,
		Scheme:   req.GetScheme(),
		Currency: req.GetCurrencyCode(),
		Data:     req.GetData(),
	})

	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"request": req,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("unable to create record"))
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusCreated, "recipient created successfully", res.ID)

}

func (c *usersController) UpdateRecipient(ctx *gin.Context) {

	srv := c.srv

	user := srv.ContextGetUser(ctx)

	recipientID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("invalid recipient_id param"))
		return
	}

	recipient, err := srv.Store.GetUserRecipient(ctx, db.GetUserRecipientParams{
		UserID: user.ID,
		ID:     recipientID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, err)
			return
		}
		srv.Logger.Error(err, map[string]interface{}{
			"request": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	scheme := recipient.Scheme
	currencyCode := recipient.Currency

	req, ok := shared.SupportedSchemeRequests[strings.ToUpper(scheme)]
	if !ok {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid scheme"))
		return
	}

	req.SetCurrency(currencyCode)
	req.SetScheme(scheme)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	v := validator.NewWithStore(ctx, srv.Store)
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	if _, err := srv.Store.UpdateRecipient(ctx, db.UpdateRecipientParams{
		UserID:   user.ID,
		ID:       recipientID,
		Scheme:   req.GetScheme(),
		Currency: req.GetCurrencyCode(),
		Data:     req.GetData(),
	}); err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"request": req,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "recipient updated successfully", nil)

}
func (c *usersController) GetRecipients(ctx *gin.Context) {

	srv := c.srv
	user := srv.ContextGetUser(ctx)

	var req domain.GetRecipientsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	v := validator.New()
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	filter := db.RecipientFilter{
		Query: strings.ToLower(req.Query),
		Filter: db.Filter{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		User: user,
	}

	res, meta, err := srv.Store.GetPaginatedRecipients(ctx, &filter)
	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"request": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "recipients retrieved successfully", gin.H{
		"recipients": res,
		"meta":       meta,
	})

}

// DeleteRecipient enables a user to remove recipient.
func (c *usersController) DeleteRecipient(ctx *gin.Context) {

	srv := c.srv

	user := srv.ContextGetUser(ctx)

	recipientID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("invalid recipient_id param"))
		return
	}

	deleteRecipientArgs := db.DeleteUserRecipientParams{
		UserID: user.ID,
		ID:     recipientID,
	}

	if err := srv.Store.DeleteUserRecipient(ctx, deleteRecipientArgs); err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"request": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "recipient removed successfully", nil)
}

func (c *usersController) GetBankInfo(ctx *gin.Context) {
	srv := c.srv

	req := struct {
		CodeName string `json:"code_name" form:"code_name" binding:"required"`
		Code     string `json:"code" form:"code" binding:"required"`
	}{}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	switch strings.ToLower(req.CodeName) {
	case "iban":
		res, err := srv.ZylaLabs.GetBankInfoThroughIBANCached(req.Code, zylalabs.CacheForever)
		if err != nil {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, errors.New("not found"))
			return
		}
		srv.SuccessJSONResponse(ctx, http.StatusOK, "OK", res)
	case "swift":
		res, err := srv.ZylaLabs.GetBankInfoThroughSwiftCached(req.Code, zylalabs.CacheForever)
		if err != nil {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, errors.New("not found"))
			return
		}
		srv.SuccessJSONResponse(ctx, http.StatusOK, "OK", res)
	case "routing_number":
		res, err := srv.ZylaLabs.GetBankInfoThroughRoutingNumberCached(req.Code, zylalabs.CacheForever)
		if err != nil {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, errors.New("not found"))
			return
		}
		srv.SuccessJSONResponse(ctx, http.StatusOK, "OK", res)

	default:
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, errors.New("unsupported code name"))
		return
	}

}

func (c *usersController) GetRecipient(ctx *gin.Context) {

	srv := c.srv
	user := srv.ContextGetUser(ctx)

	recipientID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("invalid recipient_id param"))
		return
	}

	recipient, err := srv.Store.GetUserRecipient(ctx, db.GetUserRecipientParams{
		UserID: user.ID,
		ID:     recipientID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, errors.New("recipient not found"))
			return
		}
		srv.Logger.Error(err, map[string]interface{}{
			"request": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "recipient retrieved successfully", recipient)
}
