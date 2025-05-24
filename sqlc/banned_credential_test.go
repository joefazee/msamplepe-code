package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBannedEntryQueries(t *testing.T) {

	arg := CreateBannedEntryParams{
		FieldType:  "test-field-type",
		FieldValue: "test-field-value",
		IpAddress:  "test-ip-address",
		UserAgent:  "test-user-agent",
	}
	bannedEntry, err := testQueries.CreateBannedEntry(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, bannedEntry)
	assert.NotZero(t, bannedEntry.CreatedAt)

	//Test GetBannedEntryByField
	getBannedEntry, err := testQueries.GetBannedEntryByField(context.Background(), GetBannedEntryByFieldParams{
		FieldType:  bannedEntry.FieldType,
		FieldValue: bannedEntry.FieldValue,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, getBannedEntry)
	assert.Equal(t, bannedEntry.UserAgent, getBannedEntry.UserAgent)

	// Test DeleteBannedEntry
	err = testQueries.DeleteBannedEntry(context.Background(), bannedEntry.ID)
	assert.NoError(t, err)
	getBannedEntry, err = testQueries.GetBannedEntryByField(context.Background(), GetBannedEntryByFieldParams{
		FieldType:  bannedEntry.FieldType,
		FieldValue: bannedEntry.FieldValue,
	})
	assert.Error(t, err)
	assert.Empty(t, getBannedEntry)
	assert.Equal(t, sql.ErrNoRows, err)

	// Test DeleteBannedEntryByField
	arg2 := CreateBannedEntryParams{
		FieldType:  "test-field-type2",
		FieldValue: "test-field-value2",
		IpAddress:  "test-ip-address2",
		UserAgent:  "test-user-agent2",
	}
	bannedEntry2, err := testQueries.CreateBannedEntry(context.Background(), arg2)
	assert.NoError(t, err)
	err = testQueries.DeleteBannedEntryByField(context.Background(), DeleteBannedEntryByFieldParams{
		FieldType:  bannedEntry2.FieldType,
		FieldValue: bannedEntry2.FieldValue,
	})
	assert.NoError(t, err)
	getBannedEntry, err = testQueries.GetBannedEntryByField(context.Background(), GetBannedEntryByFieldParams{
		FieldType:  bannedEntry2.FieldType,
		FieldValue: bannedEntry2.FieldValue,
	})
	assert.Error(t, err)
	assert.Empty(t, getBannedEntry)
	assert.Equal(t, sql.ErrNoRows, err)

}
