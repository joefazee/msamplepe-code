package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/core/controllers/businesses"
	"github.com/timchuks/monieverse/core/controllers/exchangerate"
	"github.com/timchuks/monieverse/core/controllers/swap"
	userCtr "github.com/timchuks/monieverse/core/controllers/users"
	"github.com/timchuks/monieverse/core/perms"
	"github.com/timchuks/monieverse/core/server"
	"github.com/timchuks/monieverse/internal/ratelimiter"
)

func registerUserRoutes(r *gin.RouterGroup, srv *server.Server) {
	user := r.Group("/")

	user.Use(srv.AuthenticatedUseRequired()).Use(srv.ActivatedUserRequired())

	user.POST("/tokens/idempotency", srv.HandleIssueIdempotencyToken())

	uctr := userCtr.NewUsersController(srv)

	user.GET("/users/profile", uctr.GetUserProfile)
	user.PATCH("/users/profile", srv.Idempotency(ratelimiter.OperationTypeUpdateUserProfile, nil),
		uctr.UpdateUserProfile)
	user.POST("/users/password", uctr.ChangePassword)

	user.POST("/users/request-for-va", uctr.RequestForVirtualDOMAccount)

	user.POST("/users/wallets/create",
		srv.Idempotency(ratelimiter.OperationTypeCreateWallet, nil),
		uctr.CreateWallet)
	user.GET("/users/wallets", uctr.GetUserWallets)

	user.GET("/users/kyc", uctr.GetUserKYC)
	user.POST("/users/uploads/identity-document", uctr.UploadIdentityDocument)

	user.GET("/users/virtual-accounts", uctr.GetUserVirtualAccounts)

	user.GET("/users/quotes", exchangerate.NewExchangeRateController(srv).GetQuotes)

	swapCtr := swap.NewSwapController(srv)
	user.POST("/users/swap",
		srv.CheckBusinessHour(),
		srv.Idempotency(ratelimiter.OperationTypeSwapCurrency, nil),
		srv.RequirePIN(), swapCtr.SwapCurrency)

	user.GET("/users/recipients", uctr.GetRecipients)
	user.POST("/users/recipients/:currency/:scheme", srv.RequirePIN(), uctr.CreateRecipient)
	user.DELETE("/users/recipients/:id", srv.RequirePIN(), uctr.DeleteRecipient)
	user.GET("/users/recipients/:id", uctr.GetRecipient)
	user.GET("/users/recipients/:id/editable", uctr.GetRecipientForEdit)
	user.PUT("/users/recipients/:id", srv.RequirePIN(), uctr.UpdateRecipient)
	user.GET("/users/recipients/supported-fields", uctr.GetCurrencySupportedFields)

	user.POST("/users/settings/set-transaction-pin", srv.CheckIfTransactionPINAlreadySet(), uctr.SetTransactionPin)
	user.PATCH("/users/settings/change-transaction-pin", uctr.ChangeTransactionPin)

	user.GET("/users/customers", uctr.GetCustomers)
	user.POST("/users/customers", uctr.CreateCustomer)
	user.GET("/users/supported-currencies", uctr.GetSupportedCurrencies)

	user.POST("/users/reset-transaction-pin", uctr.ResetTransactionPIN)
	user.POST("/users/reset-transaction-pin/complete", uctr.ResetTransactionPINComplete)

	user.POST("/transfer/new", srv.Idempotency(ratelimiter.OperationTypeCreateTransfer, nil),
		srv.RequirePIN(), uctr.CreateNewExternalTransfer)
	user.POST("/transfer/invoice", srv.RequirePIN(), uctr.UploadTransferInvoice)

	user.GET("/settings/schemes",
		uctr.GetAllPaymentSchemeConfigs)

	user.GET("/settings/schemes/:scheme",
		uctr.GetPaymentSchemeConfigs)

	user.POST("/users/referrals", srv.RequirePermission(perms.UserCanManageReferral), uctr.CreateReferral)
	user.GET("/users/referrals", srv.RequirePermission(perms.UserCanManageReferral), uctr.GetUserReferrals)
	user.DELETE("/users/referrals/:id", srv.RequirePermission(perms.UserCanManageReferral), uctr.DeleteUserReferral)

	user.GET("/transactions",
		uctr.GetUserTransactions)

	user.GET("/users/referree-status",
		uctr.GetReferreeStatus)

	user.POST("/virtual-accounts/new",
		srv.Idempotency(ratelimiter.OperationTypeNewVirtualAccount, nil),
		uctr.GetNewVirtualAccount)

	user.GET("/users/kyc/requirements", uctr.GetUserKYCRequirements)
	user.POST("/users/kyc/requirements", uctr.ProcessUserKYCRequirements)
	user.GET("/users/kyc/requirements/:id/results", uctr.GetUserKYCRequirementResults)

	user.GET("/users/recipients/get-bank-info", uctr.GetBankInfo)

	//user.GET("/run-workflow", uctr.RunWorkflow)

	biz := user.Group("/businesses")
	biz.Use(srv.CheckIsABusinessAccount())

	businessCtr := businesses.NewBusinessesController(srv)
	biz.POST("/kyb", businessCtr.CreateBusinessKYB)
	biz.PATCH("/kyb", businessCtr.UpdateBusinessDetails)
	biz.GET("/kyb", businessCtr.GetBusinessCreatedByUser)
	biz.POST("/owners", businessCtr.CreateBusinessOwner)
	biz.GET("/details", businessCtr.GetBusinessDetails)
	biz.GET("/owners", businessCtr.GetBusinessOwners)
	biz.GET("/owners/:id", businessCtr.GetBusinessOwnerByID)
	biz.PATCH("/owners/:id", businessCtr.UpdateBusinessOwnerByID)
	biz.PATCH("/kyb/:id/documents", businessCtr.UpdateBusinessDocumentByID)

	registerAdminRoutes(srv, user)

}
