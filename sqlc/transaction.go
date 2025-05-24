package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type AfterTransactionUpdateFunc func(tx Transaction) error

type UpdateTransactionTxParams struct {
	UpdateTransactionParams
}

type UpdateTransactionTxResult struct {
	Transaction Transaction
}

// UpdateTransactionTx updates a transaction and calls the afterUpdate function
// after the transaction has been updated, but before the transaction is committed.
func (store *SQLStore) UpdateTransactionTx(ctx context.Context, arg UpdateTransactionTxParams, afterUpdate AfterTransactionUpdateFunc) (UpdateTransactionTxResult, error) {

	var result UpdateTransactionTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transaction, err = q.UpdateTransaction(ctx, arg.UpdateTransactionParams)
		if err != nil {
			return err
		}

		if afterUpdate != nil {
			return afterUpdate(result.Transaction)
		}
		return nil
	})

	return result, err
}

const getTransactionRecipientData = `-- name: GetTransactionRecipientDataByID :one
SELECT
    id,
    payload->'recipient'->>'data' AS recipient_data,
    payload->'recipient'->>'id' AS recipient_id,
    payload->'recipient'->>'scheme' AS scheme,
    payload->'recipient'->>'currency' AS currency,
    status, action
FROM
    transactions
WHERE
    id = $1 AND action = 'ext-transfer'
`

type TransactionRecipientData struct {
	ID            uuid.UUID
	RecipientData string `json:"recipient_data"`
	RecipientID   string `json:"recipient_id"`
	Scheme        string `json:"scheme"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	Action        string `json:"action"`
}

// GetTransactionRecipientDataByID returns the recipient data of a transaction
// only work for transaction with action 'ext-transfer'
func (store *SQLStore) GetTransactionRecipientDataByID(ctx context.Context, id uuid.UUID) (*TransactionRecipientData, error) {
	row := store.db.QueryRowContext(ctx, getTransactionRecipientData, id)
	var data TransactionRecipientData
	err := row.Scan(&data.ID, &data.RecipientData, &data.RecipientID, &data.Scheme, &data.Currency, &data.Status, &data.Action)
	return &data, err
}

// UpdateTransactionRecipientData updates the recipient data of a transaction
// only work for transaction with action 'ext-transfer'
func (store *SQLStore) UpdateTransactionRecipientData(ctx context.Context, newData interface{}, id uuid.UUID) error {

	bs, err := json.Marshal(newData)
	if err != nil {
		return fmt.Errorf("failed to marshal new data: %w", err)
	}

	const updateTransactionRecipientData = `
		UPDATE transactions
		SET payload = jsonb_set(
			payload,
			'{recipient,data}',
			$1::jsonb,
			false
		)
		WHERE id = $2
		  AND action = 'ext-transfer';
`

	result, err := store.db.ExecContext(ctx, updateTransactionRecipientData, string(bs), id)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no transaction found with id %s and action 'ext-transfer'", id)
	}

	return nil
}
