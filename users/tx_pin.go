package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/core/server"
	"github.com/timchuks/monieverse/internal/common"
	"github.com/timchuks/monieverse/internal/core"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/notifier"
	"github.com/timchuks/monieverse/internal/settings"
	"github.com/timchuks/monieverse/internal/token"
)

func (c *usersController) ResetTransactionPIN(ctx *gin.Context) {

	srv := c.srv
	user := srv.ContextGetUser(ctx)

	_ = srv.Store.DeleteUserTokens(ctx, db.DeleteUserTokensParams{
		UserID: user.ID,
		Scope:  token.ResetPINScope,
	})

	numGenerator := token.NewNumericGenerator(server.TokenLength)
	otpService := token.NewDBTokenService(srv.Store, numGenerator, time.Now().Add(30*time.Minute))

	otp, err := otpService.SendOTP(ctx, user.ID, token.ResetPINScope)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("error completing request. try again please"))
		return
	}

	data := map[string]interface{}{
		"Code": otp.Code,
		"Name": user.AccountName(),
	}

	srv.SendNotificationFromTemplate(
		ctx,
		notifier.NewEmailRecipient(user.Email),
		"Reset your transaction PIN", "reset-transaction-pin.html.tmpl",
		data, nil)

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, nil)

}

func (c *usersController) ResetTransactionPINComplete(ctx *gin.Context) {
	srv := c.srv
	user := srv.ContextGetUser(ctx)

	input := struct {
		Code    string `json:"code"`
		NewCode string `json:"new_code"`
	}{}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	input.Code = strings.TrimSpace(input.Code)
	input.NewCode = strings.TrimSpace(input.NewCode)

	if len(input.NewCode) < minPINLength || len(input.NewCode) > maxPINLength {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("new pin must be between %d and %d characters", minPINLength, maxPINLength))
		return
	}

	userToken, err := srv.Store.GetOneTokenForUser(ctx, user.ID, token.ResetPINScope)
	if err != nil {
		srv.Logger.Error(err, nil)
		if errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, core.ErrInvalidToken)
			return
		}
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	if !checkToken(userToken, input.Code) {
		srv.ErrorJSONResponse(ctx, http.StatusUnauthorized, core.ErrInvalidToken)
		return
	}

	if err = srv.Store.DeleteUserTokens(ctx, db.DeleteUserTokensParams{
		UserID: user.ID,
		Scope:  token.ResetPINScope,
	}); err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	var userSettings settings.UserSettings
	userSettings.SetTransactionPin(input.NewCode)

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

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, nil)
}

func checkToken(userToken db.Token, code string) bool {

	if userToken.Expiry.Before(time.Now()) {
		return false
	}

	if err := common.CheckPassword(code, userToken.Hash); err != nil {
		return false
	}

	return true
}
