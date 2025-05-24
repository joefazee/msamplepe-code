package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timchuks/monieverse/internal/domain"
)

func (c *usersController) GetUserWallets(ctx *gin.Context) {

	srv := c.srv
	user := srv.ContextGetUser(ctx)

	wallets, err := srv.Store.GetUserWallets(ctx, user.ID)
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	vNums, err := srv.Store.GetUserVirtualAccounts(ctx, user.ID)
	if err != nil {
		srv.Logger.Error(err, nil)
		return
	}

	res := domain.WalletResponses{}.Marshall(false, wallets, vNums)

	srv.SuccessJSONResponse(ctx, http.StatusOK, "wallets retrieved successfully", res)
}
