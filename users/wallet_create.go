package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/internal/core"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/validator"
	"github.com/timchuks/monieverse/internal/worker"
)

func (c *usersController) CreateWallet(ctx *gin.Context) {
	srv := c.srv
	user := srv.ContextGetUser(ctx)

	req := domain.CreateWalletRequest{
		User: user,
	}

	country, err := srv.Store.GetCountry(ctx, user.CountryCode)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, core.ErrCountryNotSet)
		return

	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	v := validator.NewWithStore(ctx, srv.Store)
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}
	_, err = srv.Store.GetUserWalletByCurrency(ctx, db.GetUserWalletByCurrencyParams{
		UserID:     user.ID,
		CurrencyID: req.CurrencyID,
	})
	if err == nil {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("wallet with the selected currency already exist"))
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	supported, err := srv.Store.IsCountryCurrencySupported(ctx, country.ID, req.CurrencyID, user.AccountType)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	if !supported {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("currency not supported in the selected country"))
		return
	}

	wallet, err := srv.WalletManager.Init(ctx, user, &req.Currency)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	err = srv.TaskDistributor.Fire(ctx, worker.TaskWalletCreated, wallet)
	if err != nil {
		srv.Logger.Error(err, nil)
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "wallet created successfully", wallet)
}
