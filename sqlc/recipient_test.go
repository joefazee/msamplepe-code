package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createRandomRecipient(t *testing.T, user User) Recipient {
	arg := CreateRecipientParams{
		UserID: user.ID,
	}

	recipient, err := testQueries.CreateRecipient(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipient)
	assert.Equal(t, arg.UserID, recipient.UserID)
	assert.NotZero(t, recipient.CreatedAt)
	assert.NotZero(t, recipient.UpdatedAt)

	return recipient
}
