package users

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/useraction"
	"github.com/timchuks/monieverse/internal/validator"
)

func (c *usersController) CreateCustomer(ctx *gin.Context) {

	srv := c.srv

	var req domain.CreateCustomerRequest
	var err error

	authUser := srv.ContextGetUser(ctx)

	if err = ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	v := validator.New()
	if !req.Validate(v) {
		srv.SendValidationError(ctx, validator.NewValidationError("Validation failed", v.Errors))
		return
	}

	customer := c.tryToFindACustomer(ctx, &req, authUser)
	if customer != nil {
		srv.ErrorJSONResponse(ctx, http.StatusConflict, errors.New("customer already exists"))
		return
	}

	customerDB, err := srv.Store.CreateCustomer(ctx, db.CreateCustomerParams{
		OwnerID: authUser.ID,
		Name:    req.Name,
		Phone:   req.Phone,
		Email:   req.Email,
	})

	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"req":      req,
			"customer": customer,
		})
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("error creating record"))
		return
	}

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeCreateCustomer, "customer created", map[string]interface{}{"result": customerDB})

	srv.SuccessJSONResponse(ctx, http.StatusCreated, "Customer created successfully", customerDB)
}

func (c *usersController) GetCustomers(ctx *gin.Context) {

	srv := c.srv

	authUser := srv.ContextGetUser(ctx)

	input := struct {
		Query string `form:"query"`
	}{}

	if err := ctx.ShouldBindQuery(&input); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	filter := db.CustomerFilter{
		Filter: db.Filter{
			PageSize: 1000, //TODO: take out the hard limit
		},
		Query: strings.ToLower(input.Query),
		Owner: authUser,
	}

	customers, meta, err := srv.Store.GetCustomers(ctx, filter)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("error fetching customers"))
		return
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "Customer created successfully", gin.H{
		"customers": customers,
		"meta":      meta,
	})
}

func (c *usersController) tryToFindACustomer(ctx context.Context, req *domain.CreateCustomerRequest, owner *db.User) *db.Customer {
	if req == nil || owner == nil {
		return nil
	}

	srv := c.srv

	customer, err := srv.Store.FindCustomerByPhone(ctx, db.FindCustomerByPhoneParams{
		OwnerID: owner.ID,
		Phone:   req.Phone,
	})

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			srv.Logger.Error(err, nil)
		}
	}

	customer, err = srv.Store.FindCustomerByEmail(ctx, db.FindCustomerByEmailParams{
		OwnerID: owner.ID,
		Email:   req.Email,
	})

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			srv.Logger.Error(err, nil)
		}
		return nil
	}

	return &customer
}

func (c *usersController) GetSupportedCurrencies(ctx *gin.Context) {

	srv := c.srv

	authUser := srv.ContextGetUser(ctx)

	var out []struct {
		Name string `json:"name"`
		Code string `json:"code"`
		ID   int32  `json:"id"`
	}

	supportedCurrencies, err := srv.Store.GetCountrySupportedCurrencies(ctx, authUser.CountryCode, authUser.AccountType)
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("error fetching supported currencies"))
		return
	}

	for _, currency := range supportedCurrencies {
		out = append(out, struct {
			Name string `json:"name"`
			Code string `json:"code"`
			ID   int32  `json:"id"`
		}{
			Name: currency.Name,
			Code: currency.Code,
			ID:   currency.ID,
		})
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, "OK", out)
}
