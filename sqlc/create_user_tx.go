package db

import "context"

type AfterCreateUserFunc func(user User) error

type CreateUserTxParams struct {
	CreateUserParams
}

type CreateUserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams, afterCreate AfterCreateUserFunc) (CreateUserTxResult, error) {

	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		if afterCreate != nil {
			return afterCreate(result.User)
		}
		return nil
	})

	return result, err
}
