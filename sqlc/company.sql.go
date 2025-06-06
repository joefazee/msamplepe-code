// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: company.sql

package db

import (
	"context"
	"database/sql"
)

const addToCompanies = `-- name: AddToCompanies :one
INSERT INTO companies (short_name, long_name)
    VALUES ($1, $2) RETURNING id, short_name, long_name, created_at, updated_at
`

type AddToCompaniesParams struct {
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
}

func (q *Queries) AddToCompanies(ctx context.Context, arg AddToCompaniesParams) (Company, error) {
	row := q.db.QueryRowContext(ctx, addToCompanies, arg.ShortName, arg.LongName)
	var i Company
	err := row.Scan(
		&i.ID,
		&i.ShortName,
		&i.LongName,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteCompanyByShortName = `-- name: DeleteCompanyByShortName :exec
DELETE FROM companies WHERE short_name = $1
`

func (q *Queries) DeleteCompanyByShortName(ctx context.Context, shortName string) error {
	_, err := q.db.ExecContext(ctx, deleteCompanyByShortName, shortName)
	return err
}

const getCompanies = `-- name: GetCompanies :many
SELECT id, short_name, long_name, created_at, updated_at FROM companies
         ORDER BY short_name ASC
`

func (q *Queries) GetCompanies(ctx context.Context) ([]Company, error) {
	rows, err := q.db.QueryContext(ctx, getCompanies)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Company{}
	for rows.Next() {
		var i Company
		if err := rows.Scan(
			&i.ID,
			&i.ShortName,
			&i.LongName,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const getCompany = `-- name: GetCompany :one
SELECT id, short_name, long_name, created_at, updated_at FROM companies WHERE id = $1 LIMIT  1
`

func (q *Queries) GetCompany(ctx context.Context, id int64) (Company, error) {
	row := q.db.QueryRowContext(ctx, getCompany, id)
	var i Company
	err := row.Scan(
		&i.ID,
		&i.ShortName,
		&i.LongName,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCompanyByShortName = `-- name: GetCompanyByShortName :one
SELECT id, short_name, long_name, created_at, updated_at FROM companies WHERE short_name = $1 LIMIT  1
`

func (q *Queries) GetCompanyByShortName(ctx context.Context, shortName string) (Company, error) {
	row := q.db.QueryRowContext(ctx, getCompanyByShortName, shortName)
	var i Company
	err := row.Scan(
		&i.ID,
		&i.ShortName,
		&i.LongName,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateCompanyByID = `-- name: UpdateCompanyByID :one
UPDATE companies
SET short_name = COALESCE($1, short_name),
    long_name = COALESCE($2, long_name)
WHERE id = $3 RETURNING id, short_name, long_name, created_at, updated_at
`

type UpdateCompanyByIDParams struct {
	ShortName sql.NullString `json:"short_name"`
	LongName  sql.NullString `json:"long_name"`
	ID        int64          `json:"id"`
}

func (q *Queries) UpdateCompanyByID(ctx context.Context, arg UpdateCompanyByIDParams) (Company, error) {
	row := q.db.QueryRowContext(ctx, updateCompanyByID, arg.ShortName, arg.LongName, arg.ID)
	var i Company
	err := row.Scan(
		&i.ID,
		&i.ShortName,
		&i.LongName,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
