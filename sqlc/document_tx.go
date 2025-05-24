package db

import (
	"context"
)

type BeforeDocumentCreateFunc func() (*CreateDocumentParams, error)

func (store *SQLStore) CreateDocumentTx(ctx context.Context, before BeforeDocumentCreateFunc) (*Document, error) {

	var result Document
	var arg *CreateDocumentParams

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		if before != nil {
			arg, err = before()
			if err != nil {
				return err
			}
		}

		result, err = q.CreateDocument(ctx, *arg)
		if err != nil {
			return err
		}

		return nil
	})

	return &result, err
}

type BeforeDocumentUpdateFunc func() (*UpdateDocumentParams, error)

func (store *SQLStore) UpdateDocumentTx(ctx context.Context, before BeforeDocumentUpdateFunc) (*Document, error) {

	var result Document
	var arg *UpdateDocumentParams

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		if before != nil {
			arg, err = before()
			if err != nil {
				return err
			}
		}

		result, err = q.UpdateDocument(ctx, *arg)
		if err != nil {
			return err
		}

		return nil
	})

	return &result, err
}
