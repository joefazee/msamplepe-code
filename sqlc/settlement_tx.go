package db

import (
	"context"
)

type AfterCreateSettlementFunc func(settlement Settlement) error

type CreateSettlementTxParams struct {
	CreateSettlementParams
}

type CreateSettlementTxResult struct {
	Settlement Settlement
}

func (store *SQLStore) CreateSettlementTx(ctx context.Context, arg CreateSettlementTxParams, afterCreate AfterCreateSettlementFunc) (CreateSettlementTxResult, error) {

	var result CreateSettlementTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Settlement, err = q.CreateSettlement(ctx, arg.CreateSettlementParams)
		if err != nil {
			return err
		}

		if afterCreate != nil {
			return afterCreate(result.Settlement)
		}

		return nil
	})

	return result, err
}
