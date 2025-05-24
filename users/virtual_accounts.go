package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/timchuks/monieverse/core/perms"
	"github.com/timchuks/monieverse/core/server"
	"github.com/timchuks/monieverse/external/veriff"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/validator"
	"github.com/timchuks/monieverse/internal/worker"

	"github.com/gin-gonic/gin"
)

func (c *usersController) GetUserVirtualAccounts(ctx *gin.Context) {
	srv := c.srv
	user := srv.ContextGetUser(ctx)

	virtualAccounts, err := srv.Store.GetUserVirtualAccounts(ctx, user.ID)
	if err != nil {
		srv.ErrorJSONResponse(ctx, 500, err)
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "virtual accounts", virtualAccounts)
}

func (c *usersController) GetNewVirtualAccount(ctx *gin.Context) {
	srv := c.srv

	authUser := srv.ContextGetUser(ctx)

	var req domain.CreateNewVirtualAccountRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	if srv.Store.HasPermission(ctx, *authUser, perms.AdminPermission) {
		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid user id"))
			return
		}

		user, err := srv.Store.GetUser(ctx, userID)
		if err != nil {
			srv.Logger.Error(err, nil)
			srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid user id"))
			return
		}
		req.User = &user
	} else {
		req.User = authUser
	}

	v := validator.NewWithStore(ctx, srv.Store)
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError(server.ResponseValidationFailed, v.Errors))
		return
	}

	err := srv.TaskDistributor.Fire(ctx, worker.TaskGetNGNVirtualAccount, struct {
		UserID uuid.UUID `json:"user_id"`
	}{UserID: req.User.ID})

	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, fmt.Errorf("error requesting for virtual account"))
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "virtual account request submitted", nil)
}

// RequestForVirtualDOMAccount requests for a virtual account for a user
func (c *usersController) RequestForVirtualDOMAccount(ctx *gin.Context) {
	srv := c.srv
	user := srv.ContextGetUser(ctx)

	req := struct {
		WalletID uuid.UUID `json:"wallet_id"`
	}{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	wallet, err := srv.Store.GetWallet(ctx, req.WalletID)
	if err != nil || !strings.EqualFold(wallet.UserID.String(), user.ID.String()) {
		srv.ErrorJSONResponse(ctx, http.StatusForbidden, errors.New("request failed"))
		return
	}

	if !srv.WalletManager.VerifyWallet(&wallet) {
		srv.Logger.Error(errors.New("wallet verification failed"), map[string]interface{}{
			"wallet_id": wallet.ID,
			"action":    "va request",
			"user_id":   user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusForbidden, errors.New("request failed"))
		return
	}

	va, err := srv.Store.GetUserVirtualAccountByWalletIDProvider(ctx, db.GetUserVirtualAccountByWalletIDProviderParams{
		UserID:   user.ID,
		WalletID: wallet.ID,
		Provider: srv.EazyEuro.GetName(),
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("request failed"))
		return
	}

	if va.Provider != "" && va.AccountNumber != "" {
		srv.ErrorJSONResponse(ctx, http.StatusForbidden, errors.New("ACCOUNT_EXIST"))
		return
	}

	// we only allow users with a personal account to request for a virtual account
	if !strings.EqualFold(user.AccountType, domain.AccountTypePersonal) {
		srv.ErrorJSONResponse(ctx, http.StatusForbidden, fmt.Errorf("only users with a personal account can request for a virtual account"))
		return
	}

	_, err = srv.Store.GetUserVerificationDataByProvider(ctx, db.GetUserVerificationDataByProviderParams{
		Provider: veriff.ProviderName,
		UserID:   user.ID,
	})

	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"user_id": user.ID,
		})

		if errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusForbidden, fmt.Errorf("IDENTITY_NOT_VERIFIED"))
			return
		}

		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, fmt.Errorf("error checking for verification"))
		return
	}

	userMeta, err := srv.Store.GetUserMetas(ctx, user.ID)
	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"user_id": user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("error getting user meta data"))
		return
	}

	if !userMeta.IdentityVerified || userMeta.IdentityVerificationStatus != db.IdentityVerificationStatusApproved {
		srv.ErrorJSONResponse(ctx, http.StatusForbidden, fmt.Errorf("IDENTITY_NOT_VERIFIED"))
		return
	}

	if err = srv.TaskDistributor.Fire(ctx, worker.TaskProcessDOMAccountForIndividual, wallet.ID); err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"job":       worker.TaskProcessDOMAccountForIndividual,
			"wallet_id": wallet.ID,
			"user_id":   user.ID,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("failed to start the process"))
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "ok", nil)
}
