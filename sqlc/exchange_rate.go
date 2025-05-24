package db

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrExchangeRateNotFound = errors.New("exchange rate not found")

const (
	ExchangeRateTypeSell = "sell"
	ExchangeRateTypeBuy  = "buy"

	ExchangeRateSourceBinance = "binance"
	ExchangeRateSourceManual  = "manual"
)

type exchangeRateCalculator struct {
	store Store
}

type ExchangeRateCalculator interface {
	CalculateExchangeRate(ctx context.Context, baseCurrency Currency, quoteCurrency Currency, rateType string) (*CalculateExchangeRateRow, error)
	CalculateUserExchangeRate(ctx context.Context, user User, baseCurrency Currency, quoteCurrency Currency, rateType string) (*CalculateExchangeRateRow, error)
}

// NewExchangeRateCalculator creates a new instance of ExchangeRateCalculator
func NewExchangeRateCalculator(store Store) ExchangeRateCalculator {
	return &exchangeRateCalculator{store: store}
}

// CalculateExchangeRate calculates the exchange rate between two currencies
func (ex *exchangeRateCalculator) CalculateExchangeRate(ctx context.Context, baseCurrency Currency, quoteCurrency Currency, rateType string) (*CalculateExchangeRateRow, error) {

	arg := CalculateExchangeRateParams{
		BaseCurrencyID:  baseCurrency.ID,
		QuoteCurrencyID: quoteCurrency.ID,
		Type:            rateType,
	}
	res, err := ex.store.CalculateExchangeRate(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExchangeRateNotFound
		}
		return nil, err
	}

	return &res, nil
}

// CalculateUserExchangeRate calculates the exchange rate between two currencies for a user
func (ex *exchangeRateCalculator) CalculateUserExchangeRate(ctx context.Context, user User, baseCurrency Currency, quoteCurrency Currency, rateType string) (*CalculateExchangeRateRow, error) {

	res, err := ex.store.CalculateExchangeRate(ctx, CalculateExchangeRateParams{
		BaseCurrencyID:  baseCurrency.ID,
		QuoteCurrencyID: quoteCurrency.ID,
		UserID:          user.ID,
		Type:            rateType,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExchangeRateNotFound
		}
		return nil, err
	}
	return &res, nil
}

// GetRateString returns the rate as a string
func (r *CalculateExchangeRateRow) GetRateString() string {
	return r.Rate.String()
}

// StringToFloat64 converts the rate string to a float64
func (r *CalculateExchangeRateRow) StringToFloat64() float64 {
	f, err := strconv.ParseFloat(r.GetRateString(), 64)
	if err != nil {
		return 0
	}
	return f
}

const getPaginatedExchangeRate = `-- name: GetPaginatedExchangeRate :many
SELECT
	count(*) OVER() AS total_count,
    er.id,
    er.rate,
	er.spread,
    base_currency.name AS base_currency_name,
    quote_currency.name AS quote_currency_name,
    er.valid_from,
    er.valid_until,
	er.type,
	er.version,
	er.automate_rate,
	er.rate_source,
	er.base_exchange_rate,
	er.base_currency_id,
	er.quote_currency_id,
    quote_currency.name as quote_currency_name,
    quote_currency.code as quote_currency_code,
    concat(base_currency.code, '/', quote_currency.code) as currency_pair
FROM
    exchange_rates AS er
        INNER JOIN
    currencies AS base_currency ON er.base_currency_id = base_currency.id
        INNER JOIN
    currencies AS quote_currency ON er.quote_currency_id = quote_currency.id
ORDER BY
    er.id ASC
LIMIT $1 OFFSET $2
`

type GetPaginatedExchangeRateRow struct {
	ID                int32           `json:"id"`
	Rate              decimal.Decimal `json:"rate"`
	BaseCurrencyName  string          `json:"base_currency_name"`
	QuoteCurrencyName string          `json:"quote_currency_name"`
	ValidFrom         sql.NullTime    `json:"valid_from"`
	ValidUntil        sql.NullTime    `json:"valid_until"`
	BaseCurrencyCode  string          `json:"base_currency_code"`
	QuoteCurrencyCode string          `json:"quote_currency_code"`
	CurrencyPair      string          `json:"currency_pair"`
	Spread            decimal.Decimal `json:"spread"`
	Type              string          `json:"type"`
	Version           time.Time       `json:"version"`
	BaseExchangeRate  int32           `json:"base_exchange_rate"`
	AutomateRate      bool            `json:"automate_rate"`
	RateSource        string          `json:"rate_source"`
	BaseCurrencyID    int32           `json:"base_currency_id"`
	QuoteCurrencyID   int32           `json:"quote_currency_id"`
}

func (q *Queries) GetPaginatedExchangeRate(ctx context.Context, filter Filter) ([]GetPaginatedExchangeRateRow, Metadata, error) {
	rows, err := q.db.QueryContext(ctx, getPaginatedExchangeRate, filter.Limit(), filter.Offset())
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()
	var items []GetPaginatedExchangeRateRow
	totalRecords := 0
	for rows.Next() {
		var i GetPaginatedExchangeRateRow
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Rate,
			&i.Spread,
			&i.BaseCurrencyName,
			&i.QuoteCurrencyName,
			&i.ValidFrom,
			&i.ValidUntil,
			&i.Type,
			&i.Version,
			&i.AutomateRate,
			&i.RateSource,
			&i.BaseExchangeRate,
			&i.BaseCurrencyID,
			&i.QuoteCurrencyID,
			&i.BaseCurrencyCode,
			&i.QuoteCurrencyCode,
			&i.CurrencyPair,
		); err != nil {
			return nil, EmptyMetadata, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, EmptyMetadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, EmptyMetadata, err
	}

	metadata := CalculateMetadata(totalRecords, filter.Page, filter.Limit())
	return items, metadata, nil
}

const getPaginatedAccountExchangeRate = `-- name: GetPaginatedExchangeRate :many
SELECT
	count(*) OVER() AS total_count,
    er.id,
    er.rate,
	er.spread,
    base_currency.name AS base_currency_name,
    quote_currency.name AS quote_currency_name,
    er.valid_from,
    er.valid_until,
	er.type,
	er.version,
	er.automate_rate,
	er.rate_source,
	er.base_exchange_rate,
	er.base_currency_id,
	er.quote_currency_id,
    quote_currency.name as quote_currency_name,
    quote_currency.code as quote_currency_code,
    concat(base_currency.code, '/', quote_currency.code) as currency_pair
FROM
    account_level_rates AS er
        INNER JOIN
    currencies AS base_currency ON er.base_currency_id = base_currency.id
        INNER JOIN
    currencies AS quote_currency ON er.quote_currency_id = quote_currency.id
WHERE
	user_id = $1
ORDER BY
    er.id ASC
LIMIT $2 OFFSET $3
`

type GetPaginatedAccountExchangeRateRow struct {
	ID                int32           `json:"id"`
	Rate              decimal.Decimal `json:"rate"`
	BaseCurrencyName  string          `json:"base_currency_name"`
	QuoteCurrencyName string          `json:"quote_currency_name"`
	ValidFrom         sql.NullTime    `json:"valid_from"`
	ValidUntil        sql.NullTime    `json:"valid_until"`
	BaseCurrencyCode  string          `json:"base_currency_code"`
	QuoteCurrencyCode string          `json:"quote_currency_code"`
	CurrencyPair      string          `json:"currency_pair"`
	Spread            decimal.Decimal `json:"spread"`
	Type              string          `json:"type"`
	Version           time.Time       `json:"version"`
	BaseExchangeRate  int32           `json:"base_exchange_rate"`
	AutomateRate      bool            `json:"automate_rate"`
	RateSource        string          `json:"rate_source"`
	BaseCurrencyID    int32           `json:"base_currency_id"`
	QuoteCurrencyID   int32           `json:"quote_currency_id"`
}

func (q *Queries) GetPaginatedAccountExchangeRate(ctx context.Context, filter Filter, userID uuid.UUID) ([]GetPaginatedAccountExchangeRateRow, Metadata, error) {
	rows, err := q.db.QueryContext(ctx, getPaginatedAccountExchangeRate, userID, filter.Limit(), filter.Offset())
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()
	var items []GetPaginatedAccountExchangeRateRow
	totalRecords := 0
	for rows.Next() {
		var i GetPaginatedAccountExchangeRateRow
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Rate,
			&i.Spread,
			&i.BaseCurrencyName,
			&i.QuoteCurrencyName,
			&i.ValidFrom,
			&i.ValidUntil,
			&i.Type,
			&i.Version,
			&i.AutomateRate,
			&i.RateSource,
			&i.BaseExchangeRate,
			&i.BaseCurrencyID,
			&i.QuoteCurrencyID,
			&i.BaseCurrencyCode,
			&i.QuoteCurrencyCode,
			&i.CurrencyPair,
		); err != nil {
			return nil, EmptyMetadata, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, EmptyMetadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, EmptyMetadata, err
	}

	metadata := CalculateMetadata(totalRecords, filter.Page, filter.Limit())
	return items, metadata, nil
}
