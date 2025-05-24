package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSessionQueries(t *testing.T) {
	ctx := context.Background()

	user := createRandomUser(t, "Personal")

	sessionParams := CreateSessionParams{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: "sample-refresh-token",
		UserAgent:    "Test User Agent",
		IpAddress:    "127.0.0.1",
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Hour * 24),
	}

	session, err := testQueries.CreateSession(ctx, sessionParams)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	fetchedSession, err := testQueries.GetSession(ctx, session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, fetchedSession.ID)
	assert.Equal(t, session.UserID, fetchedSession.UserID)
	assert.Equal(t, session.RefreshToken, fetchedSession.RefreshToken)
	assert.Equal(t, session.UserAgent, fetchedSession.UserAgent)
	assert.Equal(t, session.IpAddress, fetchedSession.IpAddress)
	assert.Equal(t, session.IsBlocked, fetchedSession.IsBlocked)

	err = testQueries.DeleteSession(ctx, session.ID)
	assert.NoError(t, err)

	fetchedSession, err = testQueries.GetSession(ctx, session.ID)
	assert.Error(t, err)
}
