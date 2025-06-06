// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: wallet.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const createWallet = `-- name: CreateWallet :one
INSERT INTO wallets (user_id, currency_id)
    VALUES ($1, $2) RETURNING id, user_id, currency_id, balance, hash, created_at, updated_at, version, locked
`

type CreateWalletParams struct {
	UserID     uuid.UUID `json:"user_id"`
	CurrencyID int32     `json:"currency_id"`
}

func (q *Queries) CreateWallet(ctx context.Context, arg CreateWalletParams) (Wallet, error) {
	row := q.db.QueryRowContext(ctx, createWallet, arg.UserID, arg.CurrencyID)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CurrencyID,
		&i.Balance,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.Locked,
	)
	return i, err
}

const deleteWallet = `-- name: DeleteWallet :exec
DELETE FROM wallets WHERE id = $1
`

func (q *Queries) DeleteWallet(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteWallet, id)
	return err
}

const getUserWalletByCurrency = `-- name: GetUserWalletByCurrency :one
SELECT id, user_id, currency_id, balance, hash, created_at, updated_at, version, locked FROM wallets WHERE user_id = $1 AND currency_id = $2
`

type GetUserWalletByCurrencyParams struct {
	UserID     uuid.UUID `json:"user_id"`
	CurrencyID int32     `json:"currency_id"`
}

func (q *Queries) GetUserWalletByCurrency(ctx context.Context, arg GetUserWalletByCurrencyParams) (Wallet, error) {
	row := q.db.QueryRowContext(ctx, getUserWalletByCurrency, arg.UserID, arg.CurrencyID)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CurrencyID,
		&i.Balance,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.Locked,
	)
	return i, err
}

const getUserWallets = `-- name: GetUserWallets :many
SELECT w.id, w.user_id, w.currency_id, w.balance, w.hash, w.created_at, w.updated_at, w.version, w.locked, c.name as currency, c.decimal_places as currency_decimal_places
        FROM wallets w
        JOIN currencies c ON w.currency_id = c.id
        WHERE w.user_id = $1
`

type GetUserWalletsRow struct {
	ID                    uuid.UUID       `json:"id"`
	UserID                uuid.UUID       `json:"user_id"`
	CurrencyID            int32           `json:"currency_id"`
	Balance               decimal.Decimal `json:"balance"`
	Hash                  string          `json:"hash"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	Version               int32           `json:"version"`
	Locked                bool            `json:"locked"`
	Currency              string          `json:"currency"`
	CurrencyDecimalPlaces int16           `json:"currency_decimal_places"`
}

func (q *Queries) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]GetUserWalletsRow, error) {
	rows, err := q.db.QueryContext(ctx, getUserWallets, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetUserWalletsRow{}
	for rows.Next() {
		var i GetUserWalletsRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.CurrencyID,
			&i.Balance,
			&i.Hash,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Version,
			&i.Locked,
			&i.Currency,
			&i.CurrencyDecimalPlaces,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getWallet = `-- name: GetWallet :one
SELECT id, user_id, currency_id, balance, hash, created_at, updated_at, version, locked FROM wallets WHERE id = $1
`

func (q *Queries) GetWallet(ctx context.Context, id uuid.UUID) (Wallet, error) {
	row := q.db.QueryRowContext(ctx, getWallet, id)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CurrencyID,
		&i.Balance,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.Locked,
	)
	return i, err
}

const updateWalletBalance = `-- name: UpdateWalletBalance :one
UPDATE wallets SET balance = $1, version = version + 1 WHERE id = $2 RETURNING id, user_id, currency_id, balance, hash, created_at, updated_at, version, locked
`

type UpdateWalletBalanceParams struct {
	Balance decimal.Decimal `json:"balance"`
	ID      uuid.UUID       `json:"id"`
}

func (q *Queries) UpdateWalletBalance(ctx context.Context, arg UpdateWalletBalanceParams) (Wallet, error) {
	row := q.db.QueryRowContext(ctx, updateWalletBalance, arg.Balance, arg.ID)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CurrencyID,
		&i.Balance,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.Locked,
	)
	return i, err
}

const updateWalletHash = `-- name: UpdateWalletHash :one
UPDATE wallets SET hash = $1, version = version + 1 WHERE id = $2 RETURNING id, user_id, currency_id, balance, hash, created_at, updated_at, version, locked
`

type UpdateWalletHashParams struct {
	Hash string    `json:"hash"`
	ID   uuid.UUID `json:"id"`
}

func (q *Queries) UpdateWalletHash(ctx context.Context, arg UpdateWalletHashParams) (Wallet, error) {
	row := q.db.QueryRowContext(ctx, updateWalletHash, arg.Hash, arg.ID)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CurrencyID,
		&i.Balance,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.Locked,
	)
	return i, err
}
