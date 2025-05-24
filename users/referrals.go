package users

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/timchuks/monieverse/core/perms"
	"github.com/timchuks/monieverse/core/server"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"net/http"

	"github.com/timchuks/monieverse/internal/core"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/useraction"
	"github.com/timchuks/monieverse/internal/validator"
)

const (
	MaximumProcessingReferral = 5
	ProcessingReferralStatus  = "processing"
)

// CreateReferral handles the creating referrals for a user with permission to manage referrals.
func (c *usersController) CreateReferral(ctx *gin.Context) {
	srv := c.srv

	authUser := srv.ContextGetUser(ctx)

	var req domain.CreateReferralRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	v := validator.NewWithStore(ctx, srv.Store)
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	numOfExistingReferrals, err := srv.Store.GetProspectiveReferralCount(ctx, authUser.ID)
	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"request": req,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	if numOfExistingReferrals == MaximumProcessingReferral {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("you can only have a maximum of 5 processing referrals"))
		return
	}

	if _, err := srv.Store.CreateReferral(ctx, db.CreateReferralParams{
		RefereeID:   authUser.ID,
		Phone:       req.Phone,
		CountryCode: req.CountryCode,
	}); err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"request": req,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeCreateReferral, "referral created", map[string]interface{}{"req": req})

	srv.SuccessJSONResponse(ctx, http.StatusCreated, "referral added successfully", nil)
}

// GetUserReferrals handles the getting referrals for a user with permission to manage referrals.
func (c *usersController) GetUserReferrals(ctx *gin.Context) {
	srv := c.srv
	authUser := srv.ContextGetUser(ctx)

	referrals, err := srv.Store.GetUserReferrals(ctx, authUser.ID)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeGetReferral, "referral retrieved", nil)

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, referrals)
}

// DeleteUserReferral handles the deleting referrals with only processing status for a user with permission to manage referrals.
func (c *usersController) DeleteUserReferral(ctx *gin.Context) {
	srv := c.srv
	authUser := srv.ContextGetUser(ctx)

	referralID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("invalid referral_id param"))
		return
	}

	referral, err := srv.Store.GetReferralByID(ctx, referralID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("referral doesn't exist"))
			return
		}
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	if referral.Status != ProcessingReferralStatus {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("referral cannot be deleted"))
		return
	}

	if err := srv.Store.DeleteReferral(ctx, db.DeleteReferralParams{
		RefereeID: authUser.ID,
		ID:        referralID,
	}); err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return
	}

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeDeleteReferral, "referral deleted", map[string]interface{}{"referral": referral})

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, nil)
}

func (c *usersController) GetReferreeStatus(ctx *gin.Context) {
	srv := c.srv
	user := srv.ContextGetUser(ctx)

	isUserAReferree := !c.isUserAReferree(ctx, user.ID)

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, map[string]bool{"is_user_a_referree": isUserAReferree})
}

func (c *usersController) isUserAReferree(ctx *gin.Context, id uuid.UUID) bool {
	_, err := c.srv.Store.GetUserPermissionByCode(ctx, db.GetUserPermissionByCodeParams{
		UserID: id,
		Code:   perms.UserCanManageReferral,
	})
	return errors.Is(err, sql.ErrNoRows)
}
