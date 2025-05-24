package users

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/internal/common"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/validator"
)

func (c *usersController) ChangePassword(ctx *gin.Context) {

	srv := c.srv

	authUser := srv.ContextGetUser(ctx)

	req := domain.ChangePasswordRequest{
		User: authUser,
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	v := validator.New()
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	hashedPassword, err := common.HashPassword(req.Password)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	_, err = srv.Store.UpdateUser(ctx, db.UpdateUserParams{
		ID: authUser.ID,
		Password: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
	})

	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "password changed successfully", nil)
}
