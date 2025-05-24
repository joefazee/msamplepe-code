package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/timchuks/monieverse/internal/common"
)

func (store *SQLStore) GetSystemUser(email string) (*User, error) {
	if email == "" {
		email = "sys-root-m-t-v-2020-10-20@operation.monieverse.com"
	}

	var req = CreateUserParams{
		Email:        email,
		CountryCode:  "1",
		Phone:        "111000222",
		Password:     common.RandomString(150),
		FirstName:    "root",
		MiddleName:   "",
		LastName:     "user",
		AccountType:  "admin",
		BusinessName: "x",
	}

	u, err := store.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u, err = store.CreateUser(context.Background(), req)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &u, nil
}
