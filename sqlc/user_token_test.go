package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetOneTokenForUser(t *testing.T) {

	user := createRandomUser(t, "Business")

	testCases := []struct {
		name        string
		scope       string
		setup       func(t *testing.T, store Store, userID uuid.UUID)
		expectError bool
	}{
		{
			name: "non-expired-tokens",
			setup: func(t *testing.T, store Store, userID uuid.UUID) {
				expiry := time.Now().Add(time.Minute * 2)
				createRandomUserToken(t, userID, "verify_email", expiry)
			},
			expectError: false,
		},
		{
			name: "expired-tokens",
			setup: func(t *testing.T, store Store, userID uuid.UUID) {
				err := store.DeleteUserTokens(context.Background(), DeleteUserTokensParams{
					UserID: userID,
					Scope:  "verify_email",
				})
				assert.NoError(t, err)
				createRandomUserToken(t, userID, "verify_email", time.Now().Add(-time.Minute*2))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewStore(testDB, nil)
			tc.setup(t, store, user.ID)
			token, err := store.GetOneTokenForUser(context.Background(), user.ID, "verify_email")
			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NotEmpty(t, token)
				assert.NoError(t, err)
			}
		})
	}

}

func TestCanFindTheRightToken_ForUser(t *testing.T) {
	user := createRandomUser(t, "Business")
	expiry1 := time.Now().Add(1 * time.Minute)
	expiry2 := time.Now().Add(2 * time.Minute)
	token := createRandomUserToken(t, user.ID, "verify_email", expiry1)
	token2 := createRandomUserToken(t, user.ID, "verify_email", expiry2)

	store := NewStore(testDB, nil)
	userToken, err := store.GetOneTokenForUser(context.Background(), user.ID, "verify_email")
	assert.NoError(t, err)
	assert.NotEmpty(t, userToken)

	assert.NotEmpty(t, token)
	assert.NotEmpty(t, token2)

	assert.Equal(t, token2.UserID, userToken.UserID)
}
