package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/internal/core"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/settings"
	"github.com/timchuks/monieverse/internal/useraction"
)

const (
	minPINLength      = 6
	maxPINLength      = 6
	transactionPinKey = "transaction_pin"
)

func (c *usersController) SetTransactionPin(ctx *gin.Context) {
	srv := c.srv
	user := srv.ContextGetUser(ctx)

	var req struct {
		PIN string `json:"pin"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	if len(req.PIN) < minPINLength || len(req.PIN) > maxPINLength {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("pin must be between %d and %d characters", minPINLength, maxPINLength))
		return
	}

	var userSettings settings.UserSettings
	userSettings.SetTransactionPin(req.PIN)

	err := srv.Store.UpsertUserSetting(ctx, db.UpsertUserSettingParams{
		UserID: user.ID,
		Key:    transactionPinKey,
		Value:  userSettings.TransactionPin,
	})

	if err != nil {
		srv.Logger.Error(fmt.Errorf("error saving transaction pin: %w", err), map[string]interface{}{
			"userID": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeSetTransactionPin, "transaction pin set", map[string]interface{}{"result": userSettings})

	srv.SuccessJSONResponse(ctx, http.StatusOK, "transaction pin set successfully", nil)
}

func (c *usersController) ChangeTransactionPin(ctx *gin.Context) {
	srv := c.srv

	user := srv.ContextGetUser(ctx)

	var req struct {
		OldPin   string `json:"old_pin" binding:"required"`
		NewPin   string `json:"new_pin" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// validate password
	if err := srv.Authenticator.Checker.CheckCredentials(user, req.Password); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid password: %v", err))
		return
	}

	// validate if old pin is the same as the one indicated
	setting, err := srv.Store.GetUserSettings(ctx, user.ID)
	if err != nil {
		srv.Logger.Error(fmt.Errorf("error changing transaction pin: %w", err), map[string]interface{}{
			"userID": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	if !setting.ValidateTransactionPin(req.OldPin) {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("wrong old pin: %v", err))
		return
	}

	if len(req.NewPin) < minPINLength || len(req.NewPin) > maxPINLength {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("new pin must be between %d and %d characters", minPINLength, maxPINLength))
		return
	}

	var userSettings settings.UserSettings
	userSettings.SetTransactionPin(req.NewPin)

	err = srv.Store.UpdateUserSetting(ctx, db.UpdateUserSettingParams{
		UserID: user.ID,
		Key:    transactionPinKey,
		Value:  userSettings.TransactionPin,
	})

	if err != nil {
		srv.Logger.Error(fmt.Errorf("error changing transaction pin: %w", err), map[string]interface{}{
			"userID": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeChangeTransactionPin, "transaction pin changed", map[string]interface{}{"result": userSettings})

	srv.SuccessJSONResponse(ctx, http.StatusOK, "transaction pin changed successfully", nil)
}
