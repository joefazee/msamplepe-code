// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: identity_document.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createIdentityDocument = `-- name: CreateIdentityDocument :one
INSERT INTO user_identity_documents (
   user_id,
   document_type,
   document_number,
   document_path,
   bucket,
   storage
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, document_type, document_number, bucket, document_path, storage, created_at, updated_at
`

type CreateIdentityDocumentParams struct {
	UserID         uuid.UUID `json:"user_id"`
	DocumentType   string    `json:"document_type"`
	DocumentNumber string    `json:"document_number"`
	DocumentPath   string    `json:"document_path"`
	Bucket         string    `json:"bucket"`
	Storage        string    `json:"storage"`
}

func (q *Queries) CreateIdentityDocument(ctx context.Context, arg CreateIdentityDocumentParams) (UserIdentityDocument, error) {
	row := q.db.QueryRowContext(ctx, createIdentityDocument,
		arg.UserID,
		arg.DocumentType,
		arg.DocumentNumber,
		arg.DocumentPath,
		arg.Bucket,
		arg.Storage,
	)
	var i UserIdentityDocument
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.DocumentType,
		&i.DocumentNumber,
		&i.Bucket,
		&i.DocumentPath,
		&i.Storage,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteIdentityDocument = `-- name: DeleteIdentityDocument :exec
DELETE FROM user_identity_documents WHERE id = $1
`

func (q *Queries) DeleteIdentityDocument(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteIdentityDocument, id)
	return err
}

const getUserIdentityDocument = `-- name: GetUserIdentityDocument :one
SELECT id, user_id, document_type, document_number, bucket, document_path, storage, created_at, updated_at FROM user_identity_documents WHERE user_id = $1 LIMIT 1
`

func (q *Queries) GetUserIdentityDocument(ctx context.Context, userID uuid.UUID) (UserIdentityDocument, error) {
	row := q.db.QueryRowContext(ctx, getUserIdentityDocument, userID)
	var i UserIdentityDocument
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.DocumentType,
		&i.DocumentNumber,
		&i.Bucket,
		&i.DocumentPath,
		&i.Storage,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserIdentityDocuments = `-- name: GetUserIdentityDocuments :many
SELECT id, user_id, document_type, document_number, bucket, document_path, storage, created_at, updated_at FROM user_identity_documents WHERE user_id = $1
`

func (q *Queries) GetUserIdentityDocuments(ctx context.Context, userID uuid.UUID) ([]UserIdentityDocument, error) {
	rows, err := q.db.QueryContext(ctx, getUserIdentityDocuments, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UserIdentityDocument{}
	for rows.Next() {
		var i UserIdentityDocument
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.DocumentType,
			&i.DocumentNumber,
			&i.Bucket,
			&i.DocumentPath,
			&i.Storage,
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
