package db

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestQueries_CreateWallet(t *testing.T) {
	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	t.Run("valid_wallet_creation", func(t *testing.T) {
		arg := CreateWalletParams{
			UserID:     user.ID,
			CurrencyID: currency.ID,
		}
		wallet, err := testQueries.CreateWallet(context.Background(), arg)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, user.ID, wallet.UserID)
		assert.Equal(t, currency.ID, wallet.CurrencyID)
		assert.NotZero(t, wallet.CreatedAt)
		assert.NotZero(t, wallet.UpdatedAt)
	})

	t.Run("invalid_wallet_creation", func(t *testing.T) {
		arg := CreateWalletParams{
			UserID:     uuid.Nil,
			CurrencyID: 0,
		}
		wallet, err := testQueries.CreateWallet(context.Background(), arg)
		assert.Error(t, err)
		assert.Empty(t, wallet)
	})
}

func TestQueries_DeleteWallet(t *testing.T) {
	user := createRandomUser(t, "Personal") // Assuming this function is available in your code
	currency := createRandomCurrency(t)     // Assuming this function is available in your code

	arg := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency.ID,
	}
	wallet, err := testQueries.CreateWallet(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	t.Run("valid_wallet_deletion", func(t *testing.T) {
		err = testQueries.DeleteWallet(context.Background(), wallet.ID)
		assert.NoError(t, err)

		deletedWallet, err := testQueries.GetWallet(context.Background(), wallet.ID)
		assert.Error(t, err)
		assert.Empty(t, deletedWallet)
	})

}

func TestQueries_GetUserWalletByCurrency(t *testing.T) {
	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	arg := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency.ID,
	}
	wallet, err := testQueries.CreateWallet(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	t.Run("valid_wallet_retrieval", func(t *testing.T) {
		retrievedWallet, err := testQueries.GetUserWalletByCurrency(context.Background(), GetUserWalletByCurrencyParams{
			UserID:     user.ID,
			CurrencyID: currency.ID,
		})
		assert.NoError(t, err)
		assert.Equal(t, wallet.ID, retrievedWallet.ID)
		assert.Equal(t, wallet.UserID, retrievedWallet.UserID)
		assert.Equal(t, wallet.CurrencyID, retrievedWallet.CurrencyID)
	})

	t.Run("invalid_wallet_retrieval", func(t *testing.T) {
		retrievedWallet, err := testQueries.GetUserWalletByCurrency(context.Background(), GetUserWalletByCurrencyParams{
			UserID:     uuid.Nil,
			CurrencyID: 0,
		})
		assert.Error(t, err)
		assert.Empty(t, retrievedWallet)
	})
}

func TestQueries_GetUserWallets(t *testing.T) {
	user := createRandomUser(t, "Personal")
	currency1 := createRandomCurrency(t)
	currency2 := createRandomCurrency(t)

	arg1 := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency1.ID,
	}
	wallet1, err := testQueries.CreateWallet(context.Background(), arg1)
	assert.NoError(t, err)
	assert.NotNil(t, wallet1)

	arg2 := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency2.ID,
	}
	wallet2, err := testQueries.CreateWallet(context.Background(), arg2)
	assert.NoError(t, err)
	assert.NotNil(t, wallet2)

	wallets, err := testQueries.GetUserWallets(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, wallets, 2)

	var wallet1Found, wallet2Found bool
	for _, wallet := range wallets {
		if wallet.ID == wallet1.ID {
			assert.Equal(t, wallet1.UserID, wallet.UserID)
			assert.Equal(t, wallet1.CurrencyID, wallet.CurrencyID)
			wallet1Found = true
		} else if wallet.ID == wallet2.ID {
			assert.Equal(t, wallet2.UserID, wallet.UserID)
			assert.Equal(t, wallet2.CurrencyID, wallet.CurrencyID)
			wallet2Found = true
		}
	}
	assert.True(t, wallet1Found, "wallet1 not found in GetUserWallets result")
	assert.True(t, wallet2Found, "wallet2 not found in GetUserWallets result")
}

func TestQueries_UpdateWalletBalance(t *testing.T) {
	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)
	initialBalance, _ := decimal.NewFromString("100.00")

	arg := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency.ID,
	}
	wallet, err := testQueries.CreateWallet(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	wallet, err = testQueries.UpdateWalletBalance(context.Background(), UpdateWalletBalanceParams{
		Balance: initialBalance,
		ID:      wallet.ID,
	})
	assert.NoError(t, err)

	newBalance, _ := decimal.NewFromString("200.00")
	_, err = testQueries.UpdateWalletBalance(context.Background(), UpdateWalletBalanceParams{
		Balance: newBalance,
		ID:      wallet.ID,
	})
	assert.NoError(t, err)

	updatedWallet, err := testQueries.GetWallet(context.Background(), wallet.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedWallet)
	assert.Equal(t, newBalance, updatedWallet.Balance)
}

func TestQueries_UpdateWalletHash(t *testing.T) {
	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	arg := CreateWalletParams{
		UserID:     user.ID,
		CurrencyID: currency.ID,
	}
	wallet, err := testQueries.CreateWallet(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotNil(t, wallet)

	newHash := "new_wallet_hash"
	_, err = testQueries.UpdateWalletHash(context.Background(), UpdateWalletHashParams{
		Hash: newHash,
		ID:   wallet.ID,
	})
	assert.NoError(t, err)

	updatedWallet, err := testQueries.GetWallet(context.Background(), wallet.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedWallet)
	assert.Equal(t, newHash, updatedWallet.Hash)
}

func TestGetBalanceAsFloat64(t *testing.T) {
	tests := []struct {
		name     string
		balance  string
		expected float64
	}{
		{
			name:     "Valid balance",
			balance:  "1234.56",
			expected: 1234.56,
		},
		{
			name:     "Invalid balance",
			balance:  "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := decimal.NewFromString(tt.balance)
			w := Wallet{
				Balance: b,
			}
			actual := w.GetBalanceAsFloat64()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBalance2String(t *testing.T) {
	tests := []struct {
		name     string
		balance  string
		expected string
	}{
		{
			name:     "Valid balance",
			balance:  "1234.56",
			expected: "1234.56",
		},
		{
			name:     "Invalid balance",
			balance:  "invalid",
			expected: "0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := decimal.NewFromString(tt.balance)
			w := Wallet{
				Balance: b,
			}
			actual := w.Balance2String()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
