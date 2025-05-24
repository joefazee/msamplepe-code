package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/timchuks/monieverse/core/server"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/internal/core"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/notifier"
	"github.com/timchuks/monieverse/internal/useraction"
	"github.com/timchuks/monieverse/internal/validator"
)

type ExternalTransferPayload struct {
	Recipient *db.Recipient `json:"recipient"`
	Customer  *db.Customer  `json:"customer"`
	Wallet    *db.Wallet    `json:"wallet"`
}

func (s *ExternalTransferPayload) Bytes() []byte {
	bs, err := json.Marshal(s)
	if err != nil {
		return []byte("{}")
	}
	return bs
}

// CreateNewExternalTransfer create a new external transfer
func (c *usersController) CreateNewExternalTransfer(ctx *gin.Context) {

	srv := c.srv
	authUser := srv.ContextGetUser(ctx)

	req := domain.UserExternalTransferRequest{
		User: srv.ContextGetUser(ctx),
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	v := validator.NewWithStore(ctx, srv.Store)
	if !req.Validate(v, authUser) {
		srv.SendValidationError(ctx, validator.NewValidationError(server.ResponseValidationFailed, v.Errors))
		return
	}

	fee, err := srv.Store.GetSchemaPaymentFeeConfig(ctx, strings.ToLower(req.Recipient.Scheme))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		srv.Logger.Error(err, map[string]interface{}{
			"req": req,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, fmt.Errorf("error getting scheme config"))
		return
	}

	userCurrencyConfig, err := srv.Settings.GetCurrencyConfigurations(ctx, req.Wallet.CurrencyID, req.User.ID)
	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"req": req,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, fmt.Errorf("error getting user settings"))
		return
	}

	if req.Amount.LessThan(userCurrencyConfig.MinTransferAmount) {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("amount to transfer is less than minimum allowed"))
		return
	}

	if req.Amount.GreaterThan(userCurrencyConfig.MaxTransferAmount) {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("amount to transfer is greater than maximum allowed: %s", userCurrencyConfig.MaxTransferAmount.String()))
		return
	}

	totalFee := calculateTransferFee(req.Amount, fee)
	amountToTransfer := req.Amount.Add(totalFee)
	if amountToTransfer.IsZero() {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("amount to transfer is zero after we applied charges"))
		return
	}

	if req.Wallet.Balance.LessThan(amountToTransfer) {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("inssuficient funds to transfer"))
		return
	}

	pl := ExternalTransferPayload{
		Recipient: req.Recipient,
		Customer:  req.Customer,
		Wallet:    req.Wallet,
	}

	if req.Reason == "" {
		req.Reason = "external fund transfer"
	}

	args := db.CreateTransactionParams{
		Amount:           amountToTransfer,
		Type:             db.TransactionTypeDebit,
		Payload:          pl.Bytes(),
		Status:           db.TransactionStatusPending,
		PaymentMethod:    db.TransactionSourceWallet,
		Action:           db.TransactionActionExternalTransfer,
		Source:           db.TransactionSourceWallet,
		FeesAmount:       totalFee,
		FeesIsPercentage: fee.IsPercentage,
		CurrencyID:       req.Wallet.CurrencyID,
		Tag:              req.Reason,
	}

	_, err = srv.WalletManager.Debit(ctx, req.Wallet, args)

	if err != nil {

		srv.Logger.Error(fmt.Errorf("error during transfer: %w", err), map[string]interface{}{
			"wallet_id": req.WalletID,
			"user_id":   authUser.ID,
			"req":       req,
			"db_args":   args,
		})

		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, fmt.Errorf("unable to complete transaction"))
		return
	}

	if err = c.sendExternalTransferNotification(ctx, args, authUser); err != nil {
		srv.Logger.Error(err, nil)
		return
	}

	srv.AddUserActionToContext(ctx, useraction.UserActionTypeCreateExternalTransfer, fmt.Sprintf("transfer of %s %s to %s initiated successfully",
		args.Type,
		args.PaymentMethod,
		args.Amount), nil)

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, nil)
}

func (c *usersController) sendExternalTransferNotification(ctx *gin.Context, args db.CreateTransactionParams, authUser *db.User) error {
	srv := c.srv

	currency, err := srv.Store.GetCurrency(ctx, args.CurrencyID)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, core.ErrInternalServerError)
		return err
	}

	emailData := map[string]interface{}{
		"Topic":    "Transaction Notification",
		"Name":     cases.Title(language.Und).String(fmt.Sprintf("%v %v", srv.Sanitizer.StripHTML(authUser.FirstName), srv.Sanitizer.StripHTML(authUser.LastName))),
		"Text":     "You just moved money from your wallet.\n The amount moved is: ",
		"Amount":   srv.Sanitizer.StripHTML(message.NewPrinter(language.English).Sprintf("%d\n", args.Amount.IntPart())),
		"Currency": srv.Sanitizer.StripHTML(currency.Code),
	}

	srv.SendNotificationFromTemplate(ctx, notifier.NewEmailRecipient(authUser.Email), "Transaction Notification", "transaction-notification.html.tmpl", emailData, nil)

	srv.SendNotificationFromTemplate(ctx, notifier.NewEmailRecipient(srv.Config.AdminEmail), "New External Transfer", "transaction-notification.html.tmpl", emailData, nil)

	return nil
}

func calculateTransferFee(amount decimal.Decimal, fee db.SchemaPaymentFeesConfig) decimal.Decimal {
	if amount.IsZero() {
		return decimal.Zero
	}

	if fee.ID == 0 {
		return decimal.Zero
	}

	totalToCharge := fee.Amount
	if fee.IsPercentage {
		totalToCharge = amount.Mul(fee.Amount).Div(decimal.NewFromInt(100))
	}

	if totalToCharge.GreaterThan(fee.MaxAmount) {
		return fee.MaxAmount
	}
	return totalToCharge
}
