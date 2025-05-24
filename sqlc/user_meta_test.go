package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueries_SetUserMeta(t *testing.T) {

	store := NewStore(testDB, nil)

	ctx := context.Background()
	user1 := createRandomUser(t, "Personal")

	err := store.SetUserMeta(ctx, UserMetaCreateParams{
		UserID:   user1.ID,
		Key:      "identity_verified",
		Value:    "true",
		Datatype: "bool",
	})
	assert.NoError(t, err)

	err = store.SetUserMeta(ctx, UserMetaCreateParams{
		UserID:   user1.ID,
		Key:      "identity_verification_type",
		Value:    "manual",
		Datatype: "string",
	})
	assert.NoError(t, err)

	metas, err := store.GetUserMetas(ctx, user1.ID)
	assert.NoError(t, err)

	assert.Equal(t, true, metas.IdentityVerified)
	assert.Equal(t, "manual", metas.IdentityVerificationType)

}
