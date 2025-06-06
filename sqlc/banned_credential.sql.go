// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: banned_credential.sql

package db

import (
	"context"
)

const createBannedEntry = `-- name: CreateBannedEntry :one
INSERT INTO banned_credentials (field_type, field_value, ip_address, user_agent)
VALUES ($1, $2, $3, $4) RETURNING id, field_type, field_value, ip_address, user_agent, created_at, updated_at
`

type CreateBannedEntryParams struct {
	FieldType  string `json:"field_type"`
	FieldValue string `json:"field_value"`
	IpAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
}

func (q *Queries) CreateBannedEntry(ctx context.Context, arg CreateBannedEntryParams) (BannedCredential, error) {
	row := q.db.QueryRowContext(ctx, createBannedEntry,
		arg.FieldType,
		arg.FieldValue,
		arg.IpAddress,
		arg.UserAgent,
	)
	var i BannedCredential
	err := row.Scan(
		&i.ID,
		&i.FieldType,
		&i.FieldValue,
		&i.IpAddress,
		&i.UserAgent,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteBannedEntry = `-- name: DeleteBannedEntry :exec
DELETE FROM banned_credentials WHERE id = $1
`

func (q *Queries) DeleteBannedEntry(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteBannedEntry, id)
	return err
}

const deleteBannedEntryByField = `-- name: DeleteBannedEntryByField :exec
DELETE FROM banned_credentials WHERE field_type = $1 AND field_value = $2
`

type DeleteBannedEntryByFieldParams struct {
	FieldType  string `json:"field_type"`
	FieldValue string `json:"field_value"`
}

func (q *Queries) DeleteBannedEntryByField(ctx context.Context, arg DeleteBannedEntryByFieldParams) error {
	_, err := q.db.ExecContext(ctx, deleteBannedEntryByField, arg.FieldType, arg.FieldValue)
	return err
}

const getBannedEntryByField = `-- name: GetBannedEntryByField :one
SELECT id, field_type, field_value, ip_address, user_agent, created_at, updated_at FROM banned_credentials WHERE field_type = $1 AND field_value = $2
`

type GetBannedEntryByFieldParams struct {
	FieldType  string `json:"field_type"`
	FieldValue string `json:"field_value"`
}

func (q *Queries) GetBannedEntryByField(ctx context.Context, arg GetBannedEntryByFieldParams) (BannedCredential, error) {
	row := q.db.QueryRowContext(ctx, getBannedEntryByField, arg.FieldType, arg.FieldValue)
	var i BannedCredential
	err := row.Scan(
		&i.ID,
		&i.FieldType,
		&i.FieldValue,
		&i.IpAddress,
		&i.UserAgent,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
