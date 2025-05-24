package db

import (
	"context"
)

type BeforeIDDocCreateFunc func() (*CreateIdentityDocumentParams, error)

func (store *SQLStore) CreateIdentityDocumentTx(ctx context.Context, before BeforeIDDocCreateFunc) (*UserIdentityDocument, error) {

	var result UserIdentityDocument
	var arg *CreateIdentityDocumentParams

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		if before != nil {
			arg, err = before()
			if err != nil {
				return err
			}
		}

		result, err = q.CreateIdentityDocument(ctx, *arg)
		if err != nil {
			return err
		}

		return nil
	})

	return &result, err
}
