package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Country struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Code            string   `json:"code"`
	Enabled         bool     `json:"enabled"`
	DefaultCurrency Currency `json:"default_currency"`
}

var ErrCountryNotFound = errors.New("country not found")

func (q *Queries) GetCountry(ctx context.Context, identifier interface{}) (*Country, error) {
	query := "SELECT countries.id as country_id, " +
		"countries.name as country_name, countries.code as country_code, countries.enabled, " +
		"currencies.id, currencies.name, currencies.code, currencies.decimal_places," +
		"currencies.active, currencies.can_have_wallet, currencies.can_swap_from, currencies.can_swap_to," +
		"currencies.supported_payment_schemes FROM countries JOIN currencies ON " +
		"countries.default_currency_id = currencies.id WHERE "

	var args []interface{}
	switch v := identifier.(type) {
	case string:
		query += "countries.code = $1"
		args = append(args, v)
	case int:
		query += "countries.id = $1"
		args = append(args, v)
	default:
		return nil, fmt.Errorf("unsupported identifier type: %T", identifier)
	}

	var c Country

	row := q.db.QueryRowContext(ctx, query, args...)
	err := row.Scan(
		&c.ID,
		&c.Name,
		&c.Code,
		&c.Enabled,
		&c.DefaultCurrency.ID,
		&c.DefaultCurrency.Name,
		&c.DefaultCurrency.Code,
		&c.DefaultCurrency.DecimalPlaces,
		&c.DefaultCurrency.Active,
		&c.DefaultCurrency.CanHaveWallet,
		&c.DefaultCurrency.CanSwapFrom,
		&c.DefaultCurrency.CanSwapTo,
		&c.DefaultCurrency.SupportedPaymentSchemes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}
	return &c, nil
}

type CountryQueryFilter struct {
	ActiveOnly bool
}

func (q *Queries) GetCountries(ctx context.Context, queryFilter CountryQueryFilter, pagination Filter) ([]Country, error) {
	query := `
        SELECT 
            countries.id as country_id, 
            countries.name as country_name, 
            countries.code as country_code, 
            countries.enabled,
            currencies.id, 
            currencies.name, 
            currencies.code, 
            currencies.decimal_places,
            currencies.active, 
            currencies.can_have_wallet, 
            currencies.can_swap_from, 
            currencies.can_swap_to,
            currencies.supported_payment_schemes 
        FROM countries 
        JOIN currencies ON countries.default_currency_id = currencies.id
    `

	var args []interface{}
	argCount := 1

	if queryFilter.ActiveOnly {
		query += fmt.Sprintf(" AND countries.enabled = $%d", argCount)
		args = append(args, true)
		argCount++
	}

	query += " ORDER BY countries.name"

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, pagination.Limit(), pagination.Offset())

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying countries: %w", err)
	}
	defer rows.Close()

	var countries []Country

	for rows.Next() {
		var c Country
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Code,
			&c.Enabled,
			&c.DefaultCurrency.ID,
			&c.DefaultCurrency.Name,
			&c.DefaultCurrency.Code,
			&c.DefaultCurrency.DecimalPlaces,
			&c.DefaultCurrency.Active,
			&c.DefaultCurrency.CanHaveWallet,
			&c.DefaultCurrency.CanSwapFrom,
			&c.DefaultCurrency.CanSwapTo,
			&c.DefaultCurrency.SupportedPaymentSchemes,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning country row: %w", err)
		}
		countries = append(countries, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating country rows: %w", err)
	}

	return countries, nil
}

// AddSupportedCurrency adds a currency to a country's supported currencies
func (q *Queries) AddSupportedCurrency(ctx context.Context, countryID, currencyID int, accountType string) error {
	query := `
        INSERT INTO country_currencies (country_id, currency_id, account_type)
        VALUES ($1, $2, $3)
        ON CONFLICT (country_id, currency_id, account_type) DO NOTHING
    `
	_, err := q.db.ExecContext(ctx, query, countryID, currencyID, accountType)
	if err != nil {
		return fmt.Errorf("failed to add supported currency: %w", err)
	}
	return nil
}

// RemoveSupportedCurrency removes a currency from a country's supported currencies
func (q *Queries) RemoveSupportedCurrency(ctx context.Context, countryID, currencyID int, accountType string) error {
	query := `
        DELETE FROM country_currencies
        WHERE country_id = $1 AND currency_id = $2 AND account_type = $3
    `
	result, err := q.db.ExecContext(ctx, query, countryID, currencyID, accountType)
	if err != nil {
		return fmt.Errorf("failed to remove supported currency: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no matching country-currency pair found")
	}

	return nil
}

// IsCountryCurrencySupported checks if a country supports a specific currency
// It accepts either string identifiers (country code, currency code) or integer IDs
func (q *Queries) IsCountryCurrencySupported(ctx context.Context, countryIdentifier, currencyIdentifier interface{}, accountType string) (bool, error) {
	query := `
        SELECT CASE
            WHEN c.default_currency_id = curr.id THEN true
            WHEN EXISTS (
                SELECT 1 FROM country_currencies cc
                WHERE cc.country_id = c.id AND cc.currency_id = curr.id  AND cc.account_type = $1
            ) THEN true
            ELSE false
        END as is_supported
        FROM countries c
        JOIN currencies curr ON 1=1
        WHERE 1=1
    `

	args := []interface{}{accountType}
	argCount := 2

	// Handle country identifier
	switch v := countryIdentifier.(type) {
	case string:
		query += fmt.Sprintf(" AND c.code = $%d", argCount)
		args = append(args, v)
	case int, int32, int64:
		query += fmt.Sprintf(" AND c.id = $%d", argCount)
		args = append(args, v)
	default:
		return false, fmt.Errorf("unsupported country identifier type: %T", countryIdentifier)
	}
	argCount++

	// Handle currency identifier
	switch v := currencyIdentifier.(type) {
	case string:
		query += fmt.Sprintf(" AND curr.code = $%d", argCount)
		args = append(args, v)
	case int, int32, int64:
		query += fmt.Sprintf(" AND curr.id = $%d", argCount)
		args = append(args, v)
	default:
		return false, fmt.Errorf("unsupported currency identifier type: %T", currencyIdentifier)
	}

	var isSupported bool
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&isSupported)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("country or currency not found")
		}
		return false, fmt.Errorf("failed to check currency support: %w", err)
	}

	return isSupported, nil
}

// GetCountrySupportedCurrencies returns all supported currencies for a given country
func (q *Queries) GetCountrySupportedCurrencies(ctx context.Context, countryIdentifier interface{}, accountType string) ([]Currency, error) {
	// Step 1: Get the country and its default currency
	country, err := q.GetCountry(ctx, countryIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get country: %w", err)
	}

	defaultCurrency := Currency{
		ID:                      country.DefaultCurrency.ID,
		Name:                    country.DefaultCurrency.Name,
		Code:                    country.DefaultCurrency.Code,
		DecimalPlaces:           country.DefaultCurrency.DecimalPlaces,
		Active:                  country.DefaultCurrency.Active,
		CanHaveWallet:           country.DefaultCurrency.CanHaveWallet,
		CanSwapFrom:             country.DefaultCurrency.CanSwapFrom,
		CanSwapTo:               country.DefaultCurrency.CanSwapTo,
		SupportedPaymentSchemes: country.DefaultCurrency.SupportedPaymentSchemes,
	}

	query := `
        SELECT 
            c.id, c.name, c.code, c.decimal_places, c.active, 
            c.can_have_wallet, c.can_swap_from, c.can_swap_to, 
            c.supported_payment_schemes
        FROM currencies c
        JOIN country_currencies cc ON c.id = cc.currency_id
        WHERE cc.country_id = $1 AND c.id != $2 AND cc.account_type = $3
        ORDER BY c.name
    `

	rows, err := q.db.QueryContext(ctx, query, country.ID, defaultCurrency.ID, accountType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Currency{defaultCurrency}, nil
		}
		return nil, fmt.Errorf("failed to query additional supported currencies: %w", err)
	}
	defer rows.Close()

	currencies := []Currency{defaultCurrency}

	for rows.Next() {
		var c Currency
		err := rows.Scan(
			&c.ID, &c.Name, &c.Code, &c.DecimalPlaces, &c.Active,
			&c.CanHaveWallet, &c.CanSwapFrom, &c.CanSwapTo,
			&c.SupportedPaymentSchemes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan currency row: %w", err)
		}
		currencies = append(currencies, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating currency rows: %w", err)
	}

	return currencies, nil
}

// SimpleCountry represents a country with basic information
type SimpleCountry struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// GetEnabledCountries returns all enabled countries with their id, name, and code
func (q *Queries) GetEnabledCountries(ctx context.Context) ([]SimpleCountry, error) {
	query := `
        SELECT id, name, code
        FROM countries
        WHERE enabled = true
        ORDER BY name
    `

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled countries: %w", err)
	}
	defer rows.Close()

	var countries []SimpleCountry
	for rows.Next() {
		var c SimpleCountry
		err := rows.Scan(&c.ID, &c.Name, &c.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to scan country row: %w", err)
		}
		countries = append(countries, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating country rows: %w", err)
	}

	if len(countries) == 0 {
		return nil, fmt.Errorf("no enabled countries found")
	}

	return countries, nil
}
