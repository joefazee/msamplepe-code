package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVirtualAccountQueries(t *testing.T) {
	user := createRandomUser(t, "Personal")

	arg := CreateUserVirtualAccountNumberParams{
		Provider:      "random-provider",
		UserID:        user.ID,
		WalletID:      user.ID,
		AccountNumber: "1234567890",
		AccountName:   "monieverse limited",
		ExternalID:    "1234",
		BankName:      "monieverse bank",
		BankCode:      "12345",
		BankSlug:      "random",
		CurrencyCode:  "NGN",
	}
	virtualAccount, err := testQueries.CreateUserVirtualAccountNumber(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, virtualAccount)
	assert.Equal(t, user.ID, virtualAccount.UserID)
	assert.NotZero(t, virtualAccount.CreatedAt)
	assert.Equal(t, user.ID, virtualAccount.UserID)
	assert.Equal(t, arg.AccountNumber, virtualAccount.AccountNumber)
	assert.Equal(t, arg.Provider, virtualAccount.Provider)

	getVirtualAccount, err := testQueries.GetUserVirtualAccounts(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, getVirtualAccount)
}
