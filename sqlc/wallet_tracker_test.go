package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestWalletHistory(t *testing.T, ctx context.Context, store *SQLStore, params CreateWalletHistoryParams) WalletHistory {
	history, err := store.CreateWalletHistoryTx(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, history)
	return *history
}

func TestSQLStore_CreateWalletHistoryTx(t *testing.T) {

	store := NewStore(testDB, nil)
	require.NotNil(t, store, "testStore is not initialized")

	ctx := context.Background()

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	walletID := wallet.ID

	t.Run("CreateWithStatusNew", func(t *testing.T) {
		arg := CreateWalletHistoryParams{
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(100.00),
			NewBalance:      decimal.NewFromFloat(90.00),
			ActionPerformed: "Debit for purchase",
			Status:          WalletHistoryStatusNew,
		}

		history, err := store.CreateWalletHistoryTx(ctx, arg)
		assert.NoError(t, err)
		assert.NotNil(t, history)

		assert.NotEqual(t, int64(0), history.ID)
		assert.Equal(t, arg.WalletID, history.WalletID)
		assert.True(t, arg.OldBalance.Equal(history.OldBalance))
		assert.True(t, arg.NewBalance.Equal(history.NewBalance))
		assert.Equal(t, arg.ActionPerformed, history.ActionPerformed)
		assert.Equal(t, WalletHistoryStatusNew, history.Status)
		assert.Empty(t, history.Hash, "Hash should be empty for status 'new'")
		assert.False(t, history.CreatedAt.IsZero())
		assert.False(t, history.UpdatedAt.IsZero())
	})

	t.Run("CreateWithStatusCompleted", func(t *testing.T) {
		arg := CreateWalletHistoryParams{
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(50.00),
			NewBalance:      decimal.NewFromFloat(75.00),
			ActionPerformed: "Credit from refund",
			Status:          WalletHistoryStatusCompleted,
		}

		history, err := store.CreateWalletHistoryTx(ctx, arg)
		assert.NoError(t, err)
		assert.NotNil(t, history)

		assert.NotEqual(t, int64(0), history.ID)
		assert.Equal(t, arg.WalletID, history.WalletID)
		assert.Equal(t, WalletHistoryStatusCompleted, history.Status)
		assert.NotEmpty(t, history.Hash, "Hash should not be empty for status 'completed'")
		assert.False(t, history.CreatedAt.IsZero())
		assert.False(t, history.UpdatedAt.IsZero())

		// Verify hash
		expectedHash, hashErr := generateWalletHistoryHash(history)
		assert.NoError(t, hashErr)
		assert.Equal(t, expectedHash, history.Hash)
	})

	t.Run("CreateWithStatusFailed", func(t *testing.T) {
		arg := CreateWalletHistoryParams{
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(20.00),
			NewBalance:      decimal.NewFromFloat(20.00), // Balance might not change on failure
			ActionPerformed: "Failed withdrawal attempt",
			Status:          WalletHistoryStatusFailed,
		}

		history, err := store.CreateWalletHistoryTx(ctx, arg)
		assert.NoError(t, err)
		assert.NotNil(t, history)
		assert.Equal(t, WalletHistoryStatusFailed, history.Status)
		assert.NotEmpty(t, history.Hash, "Hash should not be empty for status 'failed'")

		expectedHash, hashErr := generateWalletHistoryHash(history)
		assert.NoError(t, hashErr)
		assert.Equal(t, expectedHash, history.Hash)
	})
}

func TestSQLStore_UpdateWalletHistoryStatusTx(t *testing.T) {

	store := &SQLStore{
		db:      testDB,
		Queries: testQueries,
	}
	assert.NotNil(t, store, "testStore is not initialized")

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	walletID := wallet.ID
	ctx := context.Background()

	// 1. Create an initial record with 'new' status
	initialParams := CreateWalletHistoryParams{
		WalletID:        walletID,
		OldBalance:      decimal.NewFromFloat(200.00),
		NewBalance:      decimal.NewFromFloat(150.00),
		ActionPerformed: "Initial transaction",
		Status:          WalletHistoryStatusNew,
	}
	initialHistory := createTestWalletHistory(t, ctx, store, initialParams)
	require.Empty(t, initialHistory.Hash)
	originalCreatedAt := initialHistory.CreatedAt
	originalUpdatedAt := initialHistory.UpdatedAt

	t.Run("UpdateStatusNewToCompleted", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt will be different if changed
		updateArg := UpdateWalletHistoryStatusParams{
			ID:        initialHistory.ID,
			NewStatus: WalletHistoryStatusCompleted,
		}
		updatedHistory, err := store.UpdateWalletHistoryStatusTx(ctx, updateArg)
		assert.NoError(t, err)
		assert.NotNil(t, updatedHistory)

		assert.Equal(t, initialHistory.ID, updatedHistory.ID)
		assert.Equal(t, WalletHistoryStatusCompleted, updatedHistory.Status)
		assert.NotEmpty(t, updatedHistory.Hash)
		assert.True(t, updatedHistory.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be more recent")
		assert.Equal(t, originalCreatedAt.UnixNano(), updatedHistory.CreatedAt.UnixNano(), "CreatedAt should not change")

		// Verify hash (it's based on original CreatedAt and new Status)
		expectedHash, hashErr := generateWalletHistoryHash(updatedHistory) // generate uses updatedHistory.Status and updatedHistory.CreatedAt
		assert.NoError(t, hashErr)
		assert.Equal(t, expectedHash, updatedHistory.Hash)

		// Update original for next test
		originalUpdatedAt = updatedHistory.UpdatedAt
	})

	t.Run("UpdateStatusCompletedToFailed", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond)
		updateArg := UpdateWalletHistoryStatusParams{
			ID:        initialHistory.ID,
			NewStatus: WalletHistoryStatusFailed,
		}
		failedHistory, err := store.UpdateWalletHistoryStatusTx(ctx, updateArg)
		assert.NoError(t, err)
		assert.NotNil(t, failedHistory)

		assert.Equal(t, WalletHistoryStatusFailed, failedHistory.Status)
		assert.NotEmpty(t, failedHistory.Hash)
		assert.NotEqual(t, initialHistory.Hash, failedHistory.Hash, "Hash should change when status changes")
		assert.True(t, failedHistory.UpdatedAt.After(originalUpdatedAt))
		assert.Equal(t, originalCreatedAt.UnixNano(), failedHistory.CreatedAt.UnixNano(), "CreatedAt should not change")

		expectedHash, hashErr := generateWalletHistoryHash(failedHistory)
		assert.NoError(t, hashErr)
		assert.Equal(t, expectedHash, failedHistory.Hash)
	})

	t.Run("UpdateNonExistentRecord", func(t *testing.T) {
		nonExistentID := int64(9999999)
		updateArg := UpdateWalletHistoryStatusParams{
			ID:        nonExistentID,
			NewStatus: WalletHistoryStatusCompleted,
		}
		_, err := store.UpdateWalletHistoryStatusTx(ctx, updateArg)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrWalletHistoryNotFound) || containsError(err, "failed to get wallet history"), "Expected ErrWalletHistoryNotFound or a wrapped version of it")
	})
}

func TestQueries_GetWalletHistory(t *testing.T) {

	store := &SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	assert.NotNil(t, store, "testStore is not initialized")

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	walletID := wallet.ID
	ctx := context.Background()

	params := CreateWalletHistoryParams{
		WalletID:        walletID,
		OldBalance:      decimal.NewFromFloat(10.00),
		NewBalance:      decimal.NewFromFloat(5.00),
		ActionPerformed: "Get test",
		Status:          WalletHistoryStatusCompleted,
	}
	createdHistory := createTestWalletHistory(t, ctx, store, params)

	t.Run("GetExistingRecord", func(t *testing.T) {
		fetchedHistory, err := store.GetWalletHistory(ctx, createdHistory.ID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedHistory)
		assert.Equal(t, createdHistory.ID, fetchedHistory.ID)
		assert.Equal(t, createdHistory.WalletID, fetchedHistory.WalletID)
		assert.True(t, createdHistory.OldBalance.Equal(fetchedHistory.OldBalance))
		assert.True(t, createdHistory.NewBalance.Equal(fetchedHistory.NewBalance))
		assert.Equal(t, createdHistory.ActionPerformed, fetchedHistory.ActionPerformed)
		assert.Equal(t, createdHistory.Status, fetchedHistory.Status)
		assert.Equal(t, createdHistory.Hash, fetchedHistory.Hash)
		assert.Equal(t, createdHistory.CreatedAt.UnixNano(), fetchedHistory.CreatedAt.UnixNano())
		assert.Equal(t, createdHistory.UpdatedAt.UnixNano(), fetchedHistory.UpdatedAt.UnixNano())
	})

	t.Run("GetNonExistentRecord", func(t *testing.T) {
		nonExistentID := int64(8888888)
		_, err := store.GetWalletHistory(ctx, nonExistentID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrWalletHistoryNotFound))
	})
}

func TestQueries_VerifyWalletHistory(t *testing.T) {

	store := &SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	assert.NotNil(t, store, "testStore is not initialized")

	user := createRandomUser(t, "Personal")
	currency := createRandomCurrency(t)

	wallet := createRandomWallet(t, user.ID, currency.ID)
	assert.NotNil(t, wallet)

	walletID := wallet.ID

	ctx := context.Background()

	t.Run("VerifyStatusNewRecord", func(t *testing.T) {
		paramsNew := CreateWalletHistoryParams{
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(30.00),
			NewBalance:      decimal.NewFromFloat(25.00),
			ActionPerformed: "Verify new",
			Status:          WalletHistoryStatusNew,
		}
		newHistory := createTestWalletHistory(t, ctx, store, paramsNew)

		isValid, err := store.Queries.VerifyWalletHistory(ctx, newHistory.ID)
		assert.NoError(t, err)
		assert.True(t, isValid, "Status 'new' record should always verify as true")
	})

	t.Run("VerifyStatusCompletedRecordValidHash", func(t *testing.T) {
		paramsCompleted := CreateWalletHistoryParams{
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(40.00),
			NewBalance:      decimal.NewFromFloat(35.00),
			ActionPerformed: "Verify completed valid",
			Status:          WalletHistoryStatusCompleted,
		}
		completedHistory := createTestWalletHistory(t, ctx, store, paramsCompleted)
		require.NotEmpty(t, completedHistory.Hash)

		isValid, err := store.VerifyWalletHistory(ctx, completedHistory.ID)
		assert.NoError(t, err)
		assert.True(t, isValid, "Status 'completed' record with valid hash should verify as true")
	})

	t.Run("VerifyStatusCompletedRecordTamperedHash", func(t *testing.T) {
		paramsTamper := CreateWalletHistoryParams{
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(50.00),
			NewBalance:      decimal.NewFromFloat(45.00),
			ActionPerformed: "Verify completed tamper",
			Status:          WalletHistoryStatusCompleted,
		}
		tamperedHistory := createTestWalletHistory(t, ctx, store, paramsTamper)
		require.NotEmpty(t, tamperedHistory.Hash)

		if store.db == nil {
			t.Skip("Skipping hash tamper test as testStore.db is nil")
		}
		_, err := store.db.ExecContext(ctx, "UPDATE wallet_history SET hash = $1 WHERE id = $2", "tamperedhashvalue", tamperedHistory.ID)
		require.NoError(t, err, "Failed to tamper hash in DB")

		isValid, err := store.Queries.VerifyWalletHistory(ctx, tamperedHistory.ID)
		assert.NoError(t, err) // VerifyWalletHistory itself shouldn't error on mismatch, just return false
		assert.False(t, isValid, "Record with tampered hash should verify as false")
	})

	t.Run("VerifyNonExistentRecord", func(t *testing.T) {
		nonExistentID := int64(7777777)
		isValid, err := store.Queries.VerifyWalletHistory(ctx, nonExistentID)
		assert.Error(t, err)
		assert.False(t, isValid)
		assert.True(t, errors.Is(err, ErrWalletHistoryNotFound) || containsError(err, "failed to get wallet history"), "Expected ErrWalletHistoryNotFound or a wrapped version")
	})
}

func Test_generateWalletHistoryHash(t *testing.T) {
	t.Run("ValidHistory", func(t *testing.T) {
		history := &WalletHistory{
			ID:              1,
			WalletID:        uuid.New(),
			OldBalance:      decimal.NewFromFloat(10.0),
			NewBalance:      decimal.NewFromFloat(5.0),
			ActionPerformed: "test action",
			CreatedAt:       time.Now().UTC(),
			Status:          WalletHistoryStatusCompleted,
		}
		hash, err := generateWalletHistoryHash(history)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 hex string length
	})

	t.Run("NilHistory", func(t *testing.T) {
		_, err := generateWalletHistoryHash(nil)
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot generate hash for nil wallet history")
	})

	t.Run("ZeroID", func(t *testing.T) {
		history := &WalletHistory{ID: 0, CreatedAt: time.Now()}
		_, err := generateWalletHistoryHash(history)
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot generate hash for wallet history with zero ID")
	})

	t.Run("ZeroCreatedAt", func(t *testing.T) {
		history := &WalletHistory{ID: 1, CreatedAt: time.Time{}}
		_, err := generateWalletHistoryHash(history)
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot generate hash for wallet history with zero CreatedAt")
	})

	t.Run("Consistency", func(t *testing.T) {
		createdAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		walletID := uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")

		history1 := &WalletHistory{
			ID:              123,
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(100.25),
			NewBalance:      decimal.NewFromFloat(90.50),
			ActionPerformed: "Consistent Action",
			CreatedAt:       createdAt,
			Status:          WalletHistoryStatusCompleted,
		}
		hash1, err1 := generateWalletHistoryHash(history1)
		assert.NoError(t, err1)

		history2 := &WalletHistory{ // Identical data
			ID:              123,
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(100.25),
			NewBalance:      decimal.NewFromFloat(90.50),
			ActionPerformed: "Consistent Action",
			CreatedAt:       createdAt,
			Status:          WalletHistoryStatusCompleted,
		}
		hash2, err2 := generateWalletHistoryHash(history2)
		assert.NoError(t, err2)
		assert.Equal(t, hash1, hash2, "Hashes for identical data should be the same")

		history3 := &WalletHistory{
			ID:              123,
			WalletID:        walletID,
			OldBalance:      decimal.NewFromFloat(100.25),
			NewBalance:      decimal.NewFromFloat(90.50),
			ActionPerformed: "Consistent Action",
			CreatedAt:       createdAt,
			Status:          WalletHistoryStatusFailed, // Different status
		}
		hash3, err3 := generateWalletHistoryHash(history3)
		assert.NoError(t, err3)
		assert.NotEqual(t, hash1, hash3, "Hashes should differ if status differs")
	})
}

func containsError(err error, substr string) bool {
	if err == nil {
		return false
	}
	return assert.Contains(nil, err.Error(), substr)
}
