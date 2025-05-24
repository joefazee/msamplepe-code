package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/timchuks/monieverse/core/controllers/shared"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/fields"
	"github.com/timchuks/monieverse/internal/schemes"
)

func (c *usersController) GetCurrencySupportedFields(ctx *gin.Context) {

	srv := c.srv
	user := srv.ContextGetUser(ctx)

	paymentSchemes := map[string]map[string][]fields.Field{}

	currencies, err := srv.Store.GetCountrySupportedCurrencies(ctx, user.CountryCode, user.AccountType)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	allSchemes, err := schemes.GetSchemes(c.srv.Store)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("failed to get schemes"))
		return
	}

	for _, currency := range currencies {
		currencySchemes := map[string][]fields.Field{}
		for _, schema := range currency.GetSupportedPaymentSchemes() {
			if _, ok := allSchemes[schema]; !ok {
				continue
			}
			currencySchemes[schema] = allSchemes[schema]
		}
		if len(currencySchemes) == 0 {
			continue
		}
		paymentSchemes[currency.Code] = currencySchemes
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "fields", paymentSchemes)

}

func (c *usersController) GetRecipientForEdit(ctx *gin.Context) {

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

	paymentSchemes, err := schemes.GetSchemes(srv.Store)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("failed to get schemes"))
		return
	}

	recipientScheme, ok := paymentSchemes[recipient.Scheme]
	if !ok {
		srv.ErrorJSONResponse(ctx, http.StatusUnprocessableEntity, fmt.Errorf("invalid recipient scheme"))
		return
	}

	supportedSchemeRequest, ok := shared.SupportedSchemeRequests[strings.ToUpper(recipient.Scheme)]
	if !ok {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid scheme"))
		return
	}

	if err = json.Unmarshal(recipient.Data, &supportedSchemeRequest); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	result := fields.SetFieldValues(recipientScheme, supportedSchemeRequest)

	srv.SuccessJSONResponse(ctx, http.StatusOK, "ok", result)

}
