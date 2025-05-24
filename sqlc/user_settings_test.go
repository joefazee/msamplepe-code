package db

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/timchuks/monieverse/internal/logger"
	"github.com/timchuks/monieverse/internal/mapper"
)

func TestQueries_GetUserSettings(t *testing.T) {

	user := createRandomUser(t, "Personal")

	zeroLogger := logger.NewZeroLogger(io.Discard, logger.LevelOff, nil)

	store := NewStore(testDB, mapper.NewConfigMapper(zeroLogger))

	err := store.CreateUserSetting(context.Background(), CreateUserSettingParams{
		UserID: user.ID,
		Key:    "transaction_pin",
		Value:  "1234",
	})

	assert.NoError(t, err)

	userSettings, err := store.GetUserSettings(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, userSettings)
	assert.Equal(t, userSettings.TransactionPin, "1234")

}
