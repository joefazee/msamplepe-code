package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserTx(t *testing.T) {
	store := NewStore(testDB, nil)

	ctx := context.Background()

	createUserParams := CreateUserParams{
		Email:       "test@example.com",
		CountryCode: "US",
		Phone:       "1234567890",
		Password:    "password123",
		FirstName:   "John",
		MiddleName:  "D",
		LastName:    "Doe",
		AccountType: "Personal",
	}

	afterCreate := func(user User) error {
		if user.Email != createUserParams.Email {
			return fmt.Errorf("email mismatch after create")
		}
		return nil
	}

	result, err := store.CreateUserTx(ctx, CreateUserTxParams{CreateUserParams: createUserParams}, afterCreate)

	assert.NoError(t, err)
	assert.NotNil(t, result.User)

	user, err := store.GetUser(ctx, result.User.ID)
	assert.NoError(t, err)

	assert.Equal(t, createUserParams.Email, user.Email)
	assert.Equal(t, createUserParams.CountryCode, user.CountryCode)
	assert.Equal(t, createUserParams.Phone, user.Phone)
	assert.Equal(t, createUserParams.FirstName, user.FirstName)
	assert.Equal(t, createUserParams.MiddleName, user.MiddleName)
	assert.Equal(t, createUserParams.LastName, user.LastName)
	assert.Equal(t, createUserParams.AccountType, user.AccountType)
}
