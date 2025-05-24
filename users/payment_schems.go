package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/core/server"
)

func (c *usersController) GetAllPaymentSchemeConfigs(ctx *gin.Context) {
	srv := c.srv
	settings, err := srv.Store.GetSchemaPaymentFeeConfigs(ctx)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, settings)
}

func (c *usersController) GetPaymentSchemeConfigs(ctx *gin.Context) {
	srv := c.srv

	setting, err := srv.Store.GetSchemaPaymentFeeConfig(ctx, strings.ToLower(ctx.Param("scheme")))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, fmt.Errorf("scheme not found"))
			return
		}

		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, setting)
}
