package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/timchuks/monieverse/internal/common"
)

func TestQueries_CreateToken(t *testing.T) {

	user := createRandomUser(t, "Personal")
	fixedDate := time.Date(2023, time.April, 9, 0, 0, 0, 0, time.UTC)
	token := createRandomUserToken(t, user.ID, "verify_email", fixedDate)

	assert.True(t, fixedDate.Equal(token.Expiry))
}

func createRandomUserToken(t *testing.T, userID uuid.UUID, scope string, expiry time.Time) Token {

	code := faker.New().Int64Between(100000, 999999)
	hashCode, err := common.HashPassword(fmt.Sprintf("%d", code))
	assert.NoError(t, err)

	req := CreateTokenParams{
		Hash:     hashCode,
		UserID:   userID,
		Scope:    scope,
		Expiry:   expiry,
		Provider: "db",
	}

	token, err := testQueries.CreateToken(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, req.Hash, token.Hash)
	assert.Equal(t, req.UserID, token.UserID)
	assert.Equal(t, req.Scope, token.Scope)
	assert.NotZero(t, token.CreatedAt)
	assert.NotZero(t, req.Expiry)
	assert.Equal(t, req.Provider, token.Provider)
	return token
}

func TestQueries_FindExistingUserToken(t *testing.T) {
	user := createRandomUser(t, "Personal")

	code := faker.New().Int64Between(100000, 999999)
	hashCode, err := common.HashPassword(fmt.Sprintf("%d", code))
	assert.NoError(t, err)

	req := CreateTokenParams{
		Hash:     hashCode,
		UserID:   user.ID,
		Scope:    "verify_email",
		Expiry:   time.Now().Add(5 * time.Minute),
		Provider: "db",
	}

	token, err := testQueries.CreateToken(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	userToken, err := testQueries.FindExistingUserToken(context.Background(), FindExistingUserTokenParams{
		Email: user.Email,
		Scope: token.Scope,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, userToken)
	assert.Equal(t, token, userToken)

}
