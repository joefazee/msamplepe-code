package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WalletHistoryStatus defines the possible statuses for a wallet history record.
type WalletHistoryStatus string

const (
	// WalletHistoryStatusNew indicates a new, unprocessed history record. Hash is not critical.
	WalletHistoryStatusNew WalletHistoryStatus = "new"
	// WalletHistoryStatusCompleted indicates a successfully processed history record. Hash is critical.
	WalletHistoryStatusCompleted WalletHistoryStatus = "completed"
	// WalletHistoryStatusFailed indicates a failed history record. Hash is critical.
	WalletHistoryStatusFailed WalletHistoryStatus = "failed"
)

// WalletHistory represents a record in the wallet_history table.
type WalletHistory struct {
	ID              int64               `json:"id"`
	WalletID        uuid.UUID           `json:"wallet_id"`
	OldBalance      decimal.Decimal     `json:"old_balance"`
	NewBalance      decimal.Decimal     `json:"new_balance"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	ActionPerformed string              `json:"action_performed"`
	Hash            string              `json:"hash"`
	Status          WalletHistoryStatus `json:"status"`
}

// ErrWalletHistoryNotFound is returned when a wallet history record is not found.
var ErrWalletHistoryNotFound = errors.New("wallet history not found")

// generateWalletHistoryHash creates a SHA256 hash for a wallet history record.
// The hash is based on ID, WalletID, OldBalance, NewBalance, ActionPerformed, the original CreatedAt, and the Status
// for which the hash is being generated.
// It's crucial that these fields are in a consistent format for the hash to be verifiable.
func generateWalletHistoryHash(history *WalletHistory) (string, error) {
	if history == nil {
		return "", errors.New("cannot generate hash for nil wallet history")
	}
	if history.ID == 0 {
		return "", errors.New("cannot generate hash for wallet history with zero ID")
	}
	if history.CreatedAt.IsZero() {
		return "", errors.New("cannot generate hash for wallet history with zero CreatedAt")
	}

	// Ensure balances are formatted consistently, matching the precision in the database (e.g., 2 decimal places).
	// CreatedAt.UnixNano() provides a consistent, high-precision timestamp representation.
	// The Status included in the hash is the status for which this hash is being generated.
	data := fmt.Sprintf("%d|%s|%s|%s|%s|%d|%s",
		history.ID,
		history.WalletID.String(),
		history.OldBalance.StringFixed(2),
		history.NewBalance.StringFixed(2),
		history.ActionPerformed,
		history.CreatedAt.UnixNano(),
		history.Status,
	)

	hasher := sha256.New()
	_, err := hasher.Write([]byte(data))
	if err != nil {
		return "", fmt.Errorf("failed to write data to hasher: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// CreateWalletHistoryParams contains the parameters for creating a new wallet history record.
type CreateWalletHistoryParams struct {
	WalletID        uuid.UUID
	OldBalance      decimal.Decimal
	NewBalance      decimal.Decimal
	ActionPerformed string
	Status          WalletHistoryStatus
}

// runCreateWalletHistoryInTx is the core logic for creating a wallet history record.
// It is intended to be called within a transaction managed by execTx.
func (q *Queries) runCreateWalletHistoryInTx(ctx context.Context, arg CreateWalletHistoryParams) (*WalletHistory, error) {
	now := time.Now().UTC()
	history := &WalletHistory{
		WalletID:        arg.WalletID,
		OldBalance:      arg.OldBalance,
		NewBalance:      arg.NewBalance,
		ActionPerformed: arg.ActionPerformed,
		Status:          arg.Status,
		CreatedAt:       now,
		UpdatedAt:       now,
		Hash:            "",
	}

	insertQuery := `
		INSERT INTO wallet_history (
			wallet_id, old_balance, new_balance, action_performed, status, hash, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id, created_at, updated_at
	`
	err := q.db.QueryRowContext(ctx, insertQuery,
		history.WalletID, history.OldBalance, history.NewBalance, history.ActionPerformed,
		history.Status, history.Hash, history.CreatedAt, history.UpdatedAt,
	).Scan(&history.ID, &history.CreatedAt, &history.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert wallet history: %w", err)
	}

	if history.Status != WalletHistoryStatusNew {
		finalHash, hashErr := generateWalletHistoryHash(history)
		if hashErr != nil {
			return nil, fmt.Errorf("wallet history inserted (ID: %d) but failed to generate its hash: %w", history.ID, hashErr)
		}
		history.Hash = finalHash

		updateHashQuery := "UPDATE wallet_history SET hash = $1 WHERE id = $2"
		_, updateErr := q.db.ExecContext(ctx, updateHashQuery, history.Hash, history.ID)
		if updateErr != nil {
			return nil, fmt.Errorf("wallet history inserted (ID: %d) but failed to update its hash: %w", history.ID, updateErr)
		}
	}
	return history, nil
}

// CreateWalletHistoryTx creates a new wallet history record within a database transaction.
func (store *SQLStore) CreateWalletHistoryTx(ctx context.Context, arg CreateWalletHistoryParams) (*WalletHistory, error) {
	var resultHistory *WalletHistory
	err := store.execTx(ctx, func(q *Queries) error {
		var txErr error
		resultHistory, txErr = q.runCreateWalletHistoryInTx(ctx, arg)
		return txErr
	})
	return resultHistory, err
}

// GetWalletHistory retrieves a single wallet history record by its ID.
// This method can be called with a Queries object that is either transaction-scoped or not.
func (q *Queries) GetWalletHistory(ctx context.Context, id int64) (*WalletHistory, error) {
	query := `
		SELECT id, wallet_id, old_balance, new_balance, created_at, updated_at, action_performed, hash, status
		FROM wallet_history
		WHERE id = $1
	`
	var history WalletHistory
	err := q.db.QueryRowContext(ctx, query, id).Scan(
		&history.ID,
		&history.WalletID,
		&history.OldBalance,
		&history.NewBalance,
		&history.CreatedAt,
		&history.UpdatedAt,
		&history.ActionPerformed,
		&history.Hash,
		&history.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletHistoryNotFound
		}
		return nil, fmt.Errorf("failed to get wallet history by ID %d: %w", id, err)
	}
	return &history, nil
}

// UpdateWalletHistoryStatusParams contains parameters for updating a wallet history's status.
type UpdateWalletHistoryStatusParams struct {
	ID        int64
	NewStatus WalletHistoryStatus
}

// runUpdateWalletHistoryStatusInTx is the core logic for updating a wallet history's status and hash.
// It is intended to be called within a transaction managed by execTx.
func (q *Queries) runUpdateWalletHistoryStatusInTx(ctx context.Context, arg UpdateWalletHistoryStatusParams) (*WalletHistory, error) {
	selectQuery := `
		SELECT id, wallet_id, old_balance, new_balance, created_at, action_performed
		FROM wallet_history
		WHERE id = $1
		FOR UPDATE
	`
	var baseHistory WalletHistory
	err := q.db.QueryRowContext(ctx, selectQuery, arg.ID).Scan(
		&baseHistory.ID,
		&baseHistory.WalletID,
		&baseHistory.OldBalance,
		&baseHistory.NewBalance,
		&baseHistory.CreatedAt,
		&baseHistory.ActionPerformed,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletHistoryNotFound
		}
		return nil, fmt.Errorf("failed to get wallet history (ID %d) for status update: %w", arg.ID, err)
	}

	historyForHash := baseHistory
	historyForHash.Status = arg.NewStatus

	newHashValue := ""
	if arg.NewStatus != WalletHistoryStatusNew {
		var hashErr error
		newHashValue, hashErr = generateWalletHistoryHash(&historyForHash)
		if hashErr != nil {
			return nil, fmt.Errorf("failed to generate hash for wallet history ID %d during status update: %w", arg.ID, hashErr)
		}
	}

	newUpdatedAt := time.Now().UTC()
	updateQuery := `
		UPDATE wallet_history
		SET status = $1, hash = $2, updated_at = $3
		WHERE id = $4
		RETURNING wallet_id, old_balance, new_balance, created_at, updated_at, action_performed, status, hash
	`
	updatedHistory := WalletHistory{ID: arg.ID}
	err = q.db.QueryRowContext(ctx, updateQuery,
		arg.NewStatus, newHashValue, newUpdatedAt, arg.ID,
	).Scan(
		&updatedHistory.WalletID,
		&updatedHistory.OldBalance,
		&updatedHistory.NewBalance,
		&updatedHistory.CreatedAt,
		&updatedHistory.UpdatedAt,
		&updatedHistory.ActionPerformed,
		&updatedHistory.Status,
		&updatedHistory.Hash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet history status for ID %d: %w", arg.ID, err)
	}
	return &updatedHistory, nil
}

// UpdateWalletHistoryStatusTx updates the status of an existing wallet history record within a database transaction.
func (store *SQLStore) UpdateWalletHistoryStatusTx(ctx context.Context, arg UpdateWalletHistoryStatusParams) (*WalletHistory, error) {
	var resultHistory *WalletHistory
	err := store.execTx(ctx, func(q *Queries) error {
		var txErr error
		resultHistory, txErr = q.runUpdateWalletHistoryStatusInTx(ctx, arg)
		return txErr
	})
	return resultHistory, err
}

// VerifyWalletHistory checks if the stored hash for a wallet history record is valid.
func (q *Queries) VerifyWalletHistory(ctx context.Context, id int64) (bool, error) {
	history, err := q.GetWalletHistory(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to get wallet history for verification (ID %d): %w", id, err)
	}

	if history.Status == WalletHistoryStatusNew {
		return true, nil
	}

	expectedHash, err := generateWalletHistoryHash(history)
	if err != nil {
		return false, fmt.Errorf("failed to generate hash for verification of wallet history ID %d: %w", id, err)
	}

	if history.Hash != expectedHash {
		return false, nil
	}
	return true, nil
}
