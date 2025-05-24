package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/timchuks/monieverse/internal/common"
)

func createRandomCurrency(t *testing.T) Currency {
	arg := CreateCurrencyParams{
		Name: common.RandomString(16),
		Code: common.RandomString(3),
	}
	currency, err := testQueries.CreateCurrency(context.Background(), arg)
	assert.NoError(t, err)
	return currency
}

func TestCurrencyOperations(t *testing.T) {
	ctx := context.Background()

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	createCurrencyParams := CreateCurrencyParams{
		Name:          "Test Currency",
		Code:          strings.ToUpper(common.RandomString(3)),
		DecimalPlaces: 2,
		Active:        true,
	}

	currency, err := store.CreateCurrency(ctx, createCurrencyParams)
	assert.NoError(t, err)
	assert.NotNil(t, currency)

	retrievedCurrency, err := store.GetCurrency(ctx, currency.ID)
	assert.NoError(t, err)
	assert.Equal(t, currency, retrievedCurrency)

	currencyByCode, err := store.GetCurrencyByCode(ctx, currency.Code)
	assert.NoError(t, err)
	assert.Equal(t, currency, currencyByCode)

	currencies, err := store.GetAllCurrencies(ctx)
	assert.NoError(t, err)
	assert.True(t, len(currencies) > 0)

	activeCurrencies, err := store.GetActiveCurrencies(ctx, sql.NullBool{})
	assert.NoError(t, err)
	assert.True(t, len(activeCurrencies) > 0)

	err = store.DeleteCurrency(ctx, currency.ID)
	assert.NoError(t, err)

	_, err = store.GetCurrency(ctx, currency.ID)
	assert.Error(t, err)
}
