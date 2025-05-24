package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/validator"
)

func (c *usersController) GetUserTransactions(ctx *gin.Context) {
	srv := c.srv
	authUser := srv.ContextGetUser(ctx)

	var req domain.TransactionRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}
	v := validator.New()
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("validation failed", v.Errors))
		return
	}

	req.Normalize()

	filter := db.TransactionFilter{
		Status: req.Status,
		Type:   req.Type,
		Query:  req.Query,
		Filter: db.Filter{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		DateBetween: req.DateBetween,
		OrderBy:     req.OrderBy,
		User:        authUser,
	}

	res, m, err := srv.Store.GetPaginatedTransactions(ctx, &filter)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, fmt.Errorf("error getting transactions"))
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "success", gin.H{
		"transactions": res,
		"meta":         m,
	})

}
