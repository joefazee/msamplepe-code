package db

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestQueries_CreateTransaction(t *testing.T) {

	// Create new random user
	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)
	wallet := createRandomWallet(t, user.ID, currency.ID)

	arg := CreateTransactionParams{
		UserID:             user.ID,
		WalletID:           wallet.ID,
		Amount:             decimal.NewFromFloat32(1984.01),
		Type:               TransactionTypeCredit,
		Source:             TransactionSourceSwap,
		Status:             TransactionStatusCompleted,
		Tag:                "test",
		PaymentMethod:      TransactionPaymentMethodBankTransfer,
		CurrencyID:         currency.ID,
		FeesAmount:         decimal.NewFromFloat32(0.33),
		FeesIsPercentage:   false,
		Rate:               decimal.NewFromFloat32(0.0044),
		Payload:            []byte("{}"),
		RequiresSettlement: false,
		SettlementStatus:   SettlementStatusNone,
	}

	tx, err := testQueries.CreateTransaction(context.Background(), arg)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, arg.UserID, tx.UserID)
	assert.Equal(t, arg.WalletID, tx.WalletID)
	assert.True(t, arg.Amount.Equal(tx.Amount))
	assert.Equal(t, arg.Type, tx.Type)
	assert.Equal(t, arg.Source, tx.Source)
	assert.Equal(t, arg.Status, tx.Status)
	assert.Equal(t, arg.Tag, tx.Tag)
	assert.Equal(t, arg.PaymentMethod, tx.PaymentMethod)
	assert.Equal(t, arg.CurrencyID, tx.CurrencyID)
	assert.True(t, arg.FeesAmount.Equal(tx.FeesAmount))
	assert.Equal(t, arg.FeesIsPercentage, tx.FeesIsPercentage)
	assert.True(t, arg.Rate.Equal(tx.Rate), "expected %s, got %s", arg.Rate, tx.Rate)
	assert.Equal(t, arg.Payload, tx.Payload)
	assert.Equal(t, arg.RequiresSettlement, tx.RequiresSettlement)
	assert.Equal(t, arg.SettlementStatus, tx.SettlementStatus)
	assert.NotZero(t, tx.CreatedAt)
	assert.NotZero(t, tx.UpdatedAt)

}

func TestQueries_UpdateTransactionTx(t *testing.T) {

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)
	wallet := createRandomWallet(t, user.ID, currency.ID)

	store := NewStore(testDB, nil)

	ctx := context.Background()

	arg := CreateTransactionParams{
		UserID:             user.ID,
		WalletID:           wallet.ID,
		CurrencyID:         currency.ID,
		Amount:             decimal.NewFromFloat(100.32),
		Status:             TransactionStatusPending,
		Source:             TransactionSourceSwap,
		RequiresSettlement: false,
		Payload:            []byte("{}"),
	}
	tx, err := store.CreateTransaction(ctx, arg)

	assert.NoError(t, err)
	assert.Equal(t, arg.UserID, tx.UserID)
	assert.Equal(t, arg.WalletID, tx.WalletID)
	assert.Equal(t, arg.CurrencyID, tx.CurrencyID)
	assert.True(t, arg.Amount.Equal(tx.Amount))
	assert.Equal(t, arg.Status, tx.Status)
	assert.Equal(t, arg.Source, tx.Source)
	assert.False(t, tx.RequiresSettlement)

	updatedTx, err := store.UpdateTransactionTx(ctx, UpdateTransactionTxParams{
		UpdateTransactionParams: UpdateTransactionParams{
			Status:             TransactionStatusCompleted,
			ID:                 tx.ID,
			RequiresSettlement: true,
		},
	}, func(tx Transaction) error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, TransactionStatusCompleted, updatedTx.Transaction.Status)
	assert.True(t, updatedTx.Transaction.RequiresSettlement)
	assert.NotEqual(t, tx.Status, updatedTx.Transaction.Status)
	assert.NotEqual(t, tx.RequiresSettlement, updatedTx.Transaction.RequiresSettlement)

	_, err = store.UpdateTransactionTx(ctx, UpdateTransactionTxParams{
		UpdateTransactionParams: UpdateTransactionParams{
			Status:             TransactionStatusFailed,
			ID:                 tx.ID,
			RequiresSettlement: false,
		},
	}, func(tx Transaction) error {
		return errors.New("unable to do some other things")
	})

	assert.Error(t, err)
	assert.Equal(t, "unable to do some other things", err.Error())

	oldT, err := store.GetTransaction(ctx, updatedTx.Transaction.ID)
	assert.NoError(t, err)

	assert.Equal(t, updatedTx.Transaction.Status, oldT.Status)
	assert.Equal(t, updatedTx.Transaction.RequiresSettlement, oldT.RequiresSettlement)
	assert.Equal(t, updatedTx.Transaction.Status, oldT.Status)
}
