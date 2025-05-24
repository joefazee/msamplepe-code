package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/timchuks/monieverse/external/veriff"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/useraction"
	"github.com/timchuks/monieverse/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/internal/domain"
)

func (c *usersController) GetUserProfile(ctx *gin.Context) {

	srv := c.srv

	authUser := srv.ContextGetUser(ctx)

	var res domain.UserProfileResponse

	user, err := srv.Store.GetUser(ctx, authUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, err)
			return
		}
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	userMeta, err := srv.Store.GetUserMetas(ctx, authUser.ID)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	userSettings, err := srv.Store.GetUserSettings(ctx, authUser.ID)
	if err != nil {
		srv.Logger.Error(fmt.Errorf("error fetch user settings: %w", err), nil)
	}
	res.HasTransactionPin = userSettings.TransactionPin != ""

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeGetProfile, "profile retrieved successfully", nil)

	res.Marshal(ctx, &user, userMeta, srv.Store)
	srv.SuccessJSONResponse(ctx, http.StatusOK, "profile retrieved successfully", res)
}

func (c *usersController) UpdateUserProfile(ctx *gin.Context) {
	srv := c.srv

	authUser := srv.ContextGetUser(ctx)
	userMeta, err := srv.Store.GetUserMetas(ctx, authUser.ID)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	var req domain.UpdateProfileRequest
	req.UserID = authUser.ID
	req.AccountType = authUser.AccountType

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := c.updateUserProfile(ctx, &req)
	if err != nil {
		var e *validator.ValidationError
		switch {
		case errors.As(err, &e):
			srv.SendValidationError(ctx, e)
		default:
			srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	vData, err := srv.Store.GetUserVerificationDataByProvider(ctx, db.GetUserVerificationDataByProviderParams{
		Provider: veriff.ProviderName,
		UserID:   user.ID,
	})

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			srv.Logger.Error(err, map[string]interface{}{
				"user_id": user.ID.String(),
			})
			srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("error fetching verification data"))
			return
		}
	}

	_, err = srv.Store.UpdateUserIdentityVerificationData(ctx, db.UpdateUserIdentityVerificationDataParams{
		Address: db.NewNullString(user.Address),
		City:    db.NewNullString(user.City),
		ID:      vData.ID,
	})

	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"user_id": user.ID.String(),
		})

		if !errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("error updating verification address"))
		}
		return
	}

	var res domain.UserProfileResponse
	res.Marshal(ctx, &user, userMeta, srv.Store)
	srv.SuccessJSONResponse(ctx, http.StatusOK, "profile updated successfully", res)
}

func (c *usersController) updateUserProfile(ctx *gin.Context, req *domain.UpdateProfileRequest) (db.User, error) {

	srv := c.srv

	v := validator.NewWithStore(ctx, srv.Store)
	if !req.Validate(v) {
		return db.User{}, validator.NewValidationError("validation failed", v.Errors)
	}

	return srv.Store.UpdateUser(ctx, db.UpdateUserParams{
		ID:            req.UserID,
		FirstName:     db.NewNullString(req.FirstName),
		MiddleName:    db.NewNullString(req.MiddleName),
		LastName:      db.NewNullString(req.LastName),
		BusinessName:  db.NewNullString(req.BusinessName),
		Address:       db.NewNullString(req.Address),
		Zipcode:       db.NewNullString(req.Zipcode),
		City:          db.NewNullString(req.City),
		State:         db.NewNullString(req.State),
		BankName:      db.NewNullString(req.BankName),
		BankCode:      db.NewNullString(req.BankCode),
		AccountNumber: db.NewNullString(req.AccountNumber),
		Bvn:           db.NewNullString(req.Bvn),
	})
}
