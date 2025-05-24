package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTx(t *testing.T) {
	ctx := context.Background()

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	createRandomUser(t, "Personal")

	tx, err := store.db.BeginTx(ctx, nil)
	assert.NoError(t, err)

	txQueries := store.WithTx(tx)

	createCurrencyParams := CreateCurrencyParams{
		Name:          "Test Currency",
		Code:          "TCR",
		DecimalPlaces: 2,
		Active:        true,
	}

	currency, err := txQueries.CreateCurrency(ctx, createCurrencyParams)
	assert.NoError(t, err)
	assert.NotNil(t, currency)

	err = tx.Rollback()
	assert.NoError(t, err)

	_, err = store.GetCurrency(ctx, currency.ID)
	assert.Error(t, err)
}
