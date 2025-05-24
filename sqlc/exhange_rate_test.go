package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func createRandomAccountLevelRate(t *testing.T, store SQLStore, userID uuid.UUID, baseCurrencyID, quoteCurrencyID int32) AccountLevelRate {
	rate, _ := decimal.NewFromString("0.2")
	arg := CreateAccountLevelExchangeRateParams{
		Rate:            rate,
		UserID:          userID,
		BaseCurrencyID:  baseCurrencyID,
		QuoteCurrencyID: quoteCurrencyID,
		ValidFrom:       NewNullTime(time.Now().UTC()),
		ValidUntil:      sql.NullTime{},
		Type:            ExchangeRateTypeBuy,
	}

	exchangeRate, err := store.CreateAccountLevelExchangeRate(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, exchangeRate)
	assert.True(t, rate.Equal(exchangeRate.Rate))
	assert.Equal(t, arg.UserID, exchangeRate.UserID)
	assert.Equal(t, arg.BaseCurrencyID, exchangeRate.BaseCurrencyID)
	assert.Equal(t, arg.QuoteCurrencyID, exchangeRate.QuoteCurrencyID)
	assert.NotZero(t, exchangeRate.ValidFrom)
	assert.Zero(t, exchangeRate.ValidUntil)
	assert.Equal(t, arg.Type, exchangeRate.Type)
	assert.True(t, arg.ValidFrom.Time.Round(time.Millisecond).Equal(exchangeRate.ValidFrom.Time.Round(time.Millisecond)))
	assert.True(t, arg.ValidUntil.Time.Round(time.Millisecond).Equal(exchangeRate.ValidUntil.Time.Round(time.Millisecond)))
	return exchangeRate
}

func createRandomAgentDailyDiscount(t *testing.T, store SQLStore, userID uuid.UUID, baseCurrencyID int32) AgentDailyDiscount {

	discountAmount, _ := decimal.NewFromString("2")
	topUpAmount, _ := decimal.NewFromString("10000")

	arg := CreateAgentDailyDiscountParams{
		UserID:           userID,
		BaseCurrencyID:   baseCurrencyID,
		DiscountAmount:   discountAmount,
		TopUpAmount:      topUpAmount,
		DiscountMultiple: decimal.NewFromFloat(1000),
		StartTimestamp:   NewNullTime(time.Now().UTC()),
		EndTimestamp:     NewNullTime(time.Now().Add(24 * time.Hour).UTC()),
	}

	agentDailyDiscount, err := store.CreateAgentDailyDiscount(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, agentDailyDiscount)
	assert.Equal(t, arg.UserID, agentDailyDiscount.UserID)
	assert.Equal(t, arg.BaseCurrencyID, agentDailyDiscount.BaseCurrencyID)
	assert.True(t, arg.DiscountAmount.Equal(agentDailyDiscount.DiscountAmount))
	assert.True(t, arg.TopUpAmount.Equal(agentDailyDiscount.TopUpAmount))
	assert.NotZero(t, agentDailyDiscount.StartTimestamp)
	assert.NotZero(t, agentDailyDiscount.EndTimestamp)
	assert.True(t, arg.StartTimestamp.Time.Round(time.Millisecond).Equal(agentDailyDiscount.StartTimestamp.Time.Round(time.Millisecond)))
	assert.True(t, arg.EndTimestamp.Time.Round(time.Millisecond).Equal(agentDailyDiscount.EndTimestamp.Time.Round(time.Millisecond)))

	return agentDailyDiscount
}

func createRandomExchangeRate(t *testing.T, baseCurrencyID, quoteCurrencyID int32) ExchangeRate {
	rate, _ := decimal.NewFromString("200")
	arg := CreateExchangeRateParams{
		BaseCurrencyID:  baseCurrencyID,
		Type:            ExchangeRateTypeBuy,
		QuoteCurrencyID: quoteCurrencyID,
		Rate:            rate,
		Spread:          decimal.NewFromFloat(0.2),
		ValidFrom:       NewNullTime(time.Now()),
		ValidUntil:      sql.NullTime{},
	}
	exchangeRate, err := testQueries.CreateExchangeRate(context.Background(), arg)
	assert.NoError(t, err)
	return exchangeRate
}

func TestExchangeRateOperations(t *testing.T) {
	ctx := context.Background()

	baseCurrency := createRandomCurrency(t)
	quoteCurrency := createRandomCurrency(t)

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}
	user := createRandomUser(t, "Business")

	rate, _ := decimal.NewFromString("1.5")
	createExchangeRateParams := CreateExchangeRateParams{
		BaseCurrencyID:  baseCurrency.ID,
		QuoteCurrencyID: quoteCurrency.ID,
		Rate:            rate,
		ValidFrom:       NewNullTime(time.Now()),
		ValidUntil:      sql.NullTime{},
		Spread:          decimal.NewFromFloat(2),
		Type:            ExchangeRateTypeBuy,
	}

	exchangeRate, err := store.CreateExchangeRate(ctx, createExchangeRateParams)
	assert.NoError(t, err)
	assert.NotNil(t, exchangeRate)

	createAccountLevelExchangeRateParams := CreateAccountLevelExchangeRateParams{
		UserID:          user.ID,
		BaseCurrencyID:  baseCurrency.ID,
		QuoteCurrencyID: quoteCurrency.ID,
		Rate:            rate,
		Spread:          decimal.NewFromFloat(2),
		ValidFrom:       NewNullTime(time.Now()),
		ValidUntil:      sql.NullTime{},
		Type:            ExchangeRateTypeBuy,
	}

	accountLevelRate, err := store.CreateAccountLevelExchangeRate(ctx, createAccountLevelExchangeRateParams)
	assert.NoError(t, err)
	assert.NotNil(t, accountLevelRate)

	calculateExchangeRateParams := CalculateExchangeRateParams{
		BaseCurrencyID:  baseCurrency.ID,
		QuoteCurrencyID: quoteCurrency.ID,
		UserID:          accountLevelRate.UserID,
		Type:            ExchangeRateTypeBuy,
	}

	calculateExchangeRateRow, err := store.CalculateExchangeRate(ctx, calculateExchangeRateParams)
	assert.NoError(t, err)
	assert.NotNil(t, calculateExchangeRateRow)
	assert.True(t, rate.Equal(calculateExchangeRateRow.Rate))

	rate2, _ := decimal.NewFromString("1.6")

	updateAccountLevelExchangeRateParams := UpdateAccountLevelExchangeRateParams{
		BaseCurrencyID:  NewNullInt32(1),
		QuoteCurrencyID: NewNullInt32(2),
		Rate:            rate2,
		Spread:          decimal.NewFromFloat(2),
		ValidFrom:       NewNullTime(time.Now()),
		UserID:          NewNullUUID(accountLevelRate.UserID),
		ValidUntil:      sql.NullTime{},
		ID:              accountLevelRate.ID,
		Type:            NewNullString(ExchangeRateTypeBuy),
	}

	updatedAccountLevelRate, err := store.UpdateAccountLevelExchangeRate(ctx, updateAccountLevelExchangeRateParams)
	assert.NoError(t, err)
	assert.NotNil(t, updatedAccountLevelRate)
	assert.True(t, rate2.Equal(updatedAccountLevelRate.Rate))

	updateExchangeRateParams := UpdateExchangeRateParams{
		BaseCurrencyID:  sql.NullInt32{Int32: 1, Valid: true},
		QuoteCurrencyID: sql.NullInt32{Int32: 2, Valid: true},
		Rate:            rate2,
		ValidFrom:       NewNullTime(time.Now()),
		ValidUntil:      sql.NullTime{},
		ID:              exchangeRate.ID,
		Type:            NewNullString(ExchangeRateTypeBuy),
	}

	updatedExchangeRate, err := store.UpdateExchangeRate(ctx, updateExchangeRateParams)
	assert.NoError(t, err)
	assert.NotNil(t, updatedExchangeRate)
	assert.True(t, rate2.Equal(updatedExchangeRate.Rate))

}

func TestCalculateExchangeRate(t *testing.T) {
	ctx := context.Background()

	baseCurrency := createRandomCurrency(t)
	quoteCurrency := createRandomCurrency(t)

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	user := createRandomUser(t, "Business")

	t.Run("only exchange rate", func(t *testing.T) {

		exchangeRate := createRandomExchangeRate(t, baseCurrency.ID, quoteCurrency.ID)

		result, err := store.CalculateExchangeRate(ctx, CalculateExchangeRateParams{
			BaseCurrencyID:  baseCurrency.ID,
			QuoteCurrencyID: quoteCurrency.ID,
			UserID:          user.ID,
			Type:            ExchangeRateTypeBuy,
		})

		assert.NoError(t, err)
		assert.True(t, exchangeRate.Rate.Equal(result.Rate))
		assert.Equal(t, exchangeRate.ID, result.ExchangeRateID)
		assert.True(t, result.ExchangeRate.Equal(exchangeRate.Rate))
		assert.True(t, result.IsBasedRate)
		assert.False(t, result.HasDiscount)
		assert.True(t, result.AccountLevelRate.Equal(decimal.Zero))
		assert.True(t, result.CalculatedDiscountAmount.Equal(decimal.Zero))
		assert.NotNil(t, result.ExchangeRateValidFrom)
		assert.False(t, result.ExchangeRateValidUntil.Valid)
		assert.False(t, result.AccountLevelValidFrom.Valid)
		assert.False(t, result.AccountLevelValidUntil.Valid)
		assert.Equal(t, result.AccountLevelRateID, int32(0))

		err = store.DeleteExchangeRate(ctx, exchangeRate.ID)
		assert.NoError(t, err)
	})

	t.Run("account level rate", func(t *testing.T) {

		accountLevelRate := createRandomAccountLevelRate(t, store, user.ID, baseCurrency.ID, quoteCurrency.ID)

		result, err := store.CalculateExchangeRate(ctx, CalculateExchangeRateParams{
			BaseCurrencyID:  baseCurrency.ID,
			QuoteCurrencyID: quoteCurrency.ID,
			UserID:          user.ID,
			Type:            ExchangeRateTypeBuy,
		})

		assert.NoError(t, err)
		assert.True(t, accountLevelRate.Rate.Equal(result.Rate))
		assert.False(t, result.IsBasedRate)
		assert.True(t, accountLevelRate.Rate.Equal(result.Rate))
		assert.Equal(t, accountLevelRate.ID, result.AccountLevelRateID)
		assert.Equal(t, result.ExchangeRateID, int32(0))
		assert.False(t, result.HasDiscount)
		assert.True(t, result.ExchangeRate.Equal(decimal.Zero))
		assert.True(t, result.CalculatedDiscountAmount.Equal(decimal.Zero))
		assert.False(t, result.ExchangeRateValidFrom.Valid)
		assert.False(t, result.ExchangeRateValidUntil.Valid)
		assert.True(t, result.AccountLevelValidFrom.Valid)
		assert.False(t, result.AccountLevelValidUntil.Valid)

		err = testQueries.DeleteAccountLevelExchangeRate(ctx, accountLevelRate.ID)
		assert.NoError(t, err)
	})

	t.Run("agent daily discount on account level rate", func(t *testing.T) {

		createRandomExchangeRate(t, baseCurrency.ID, quoteCurrency.ID)
		rate, _ := decimal.NewFromString("20")

		createAccountLevelExchangeRateParams := CreateAccountLevelExchangeRateParams{
			UserID:          user.ID,
			BaseCurrencyID:  baseCurrency.ID,
			QuoteCurrencyID: quoteCurrency.ID,
			Rate:            rate,
			ValidFrom:       NewNullTime(time.Now()),
			ValidUntil:      sql.NullTime{},
			Type:            ExchangeRateTypeBuy,
		}

		accountLevelRate, err := store.CreateAccountLevelExchangeRate(ctx, createAccountLevelExchangeRateParams)
		assert.NoError(t, err)
		assert.NotNil(t, accountLevelRate)

		agentDiscount := createRandomAgentDailyDiscount(t, store, user.ID, accountLevelRate.BaseCurrencyID)

		arg := CalculateExchangeRateParams{
			BaseCurrencyID:  baseCurrency.ID,
			QuoteCurrencyID: accountLevelRate.QuoteCurrencyID,
			UserID:          user.ID,
			Type:            ExchangeRateTypeBuy,
		}
		assert.Equal(t, arg.BaseCurrencyID, accountLevelRate.BaseCurrencyID, "base currency id mismatch")
		assert.Equal(t, arg.QuoteCurrencyID, accountLevelRate.QuoteCurrencyID, "quote currency id mismatch")
		assert.Equal(t, arg.UserID, accountLevelRate.UserID, "user id mismatch")

		agentRate, err := store.CalculateExchangeRate(ctx, arg)
		assert.NoError(t, err, "error calculating agent rate")

		totalDiscount := agentDiscount.DiscountAmount.Mul(agentDiscount.TopUpAmount.Div(agentDiscount.DiscountMultiple).Floor())

		expectedRate := accountLevelRate.Rate.Sub(totalDiscount)
		assert.True(t, expectedRate.Equal(agentRate.Rate), "expected rate: %s, calculated agent rate: %s %+v", expectedRate, agentRate.Rate, agentRate)
		assert.True(t, agentRate.HasDiscount)
		assert.False(t, agentRate.IsBasedRate)

	})

	t.Run("agent daily discount exchange rate", func(t *testing.T) {

		err := store.DeleteAllExchangeRate(ctx)
		assert.NoError(t, err)

		err = store.DeleteAllAccountLevelExchangeRate(ctx)
		assert.NoError(t, err)

		err = store.DeleteAllAgentDailyDiscount(ctx)
		assert.NoError(t, err)

		exchangeRate := createRandomExchangeRate(t, baseCurrency.ID, quoteCurrency.ID)

		agentDiscount := createRandomAgentDailyDiscount(t, store, user.ID, exchangeRate.BaseCurrencyID)

		arg := CalculateExchangeRateParams{
			BaseCurrencyID:  baseCurrency.ID,
			QuoteCurrencyID: exchangeRate.QuoteCurrencyID,
			UserID:          user.ID,
			Type:            ExchangeRateTypeBuy,
		}

		assert.Equal(t, arg.BaseCurrencyID, exchangeRate.BaseCurrencyID, "base currency id mismatch")
		assert.Equal(t, arg.QuoteCurrencyID, exchangeRate.QuoteCurrencyID, "quote currency id mismatch")

		agentRate, err := store.CalculateExchangeRate(ctx, arg)
		assert.NoError(t, err, "error calculating agent rate")

		totalDiscount := agentDiscount.DiscountAmount.Mul(agentDiscount.TopUpAmount.Div(agentDiscount.DiscountMultiple).Floor())

		expectedRate := exchangeRate.Rate.Sub(totalDiscount)
		assert.True(t, expectedRate.Equal(agentRate.Rate), "expected rate: %s, calculated agent rate: %s %+v", expectedRate, agentRate.Rate, agentRate)
		assert.True(t, agentRate.HasDiscount)
		assert.True(t, agentRate.IsBasedRate)

	})
}

func TestStringToFloat64(t *testing.T) {
	testCases := []struct {
		name       string
		rateString string
		expected   float64
	}{
		{
			name:       "Valid float string",
			rateString: "123.45",
			expected:   123.45,
		},
		{
			name:       "Valid integer string",
			rateString: "42",
			expected:   42,
		},
		{
			name:       "Empty string",
			rateString: "",
			expected:   0,
		},
		{
			name:       "Invalid string",
			rateString: "invalid",
			expected:   0,
		},
		{
			name:       "Exponential notation",
			rateString: "1.23e+3",
			expected:   1230,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, _ := decimal.NewFromString(tc.rateString)
			row := &CalculateExchangeRateRow{
				Rate: d,
			}

			result := row.StringToFloat64()

			assert.True(t, d.Equal(decimal.NewFromFloat(result)))
		})
	}
}
