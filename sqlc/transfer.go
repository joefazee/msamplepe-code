package db

import (
	"context"
	"fmt"
)

func (store *SQLStore) GetPaginatedTransfers(ctx context.Context, filter *TransactionFilter) ([]TransactionRow, Metadata, error) {
	filter.WhereConditions = append(filter.WhereConditions, fmt.Sprintf("t.action = '%s'", TransactionActionExternalTransfer))
	return store.GetPaginatedTransactions(ctx, filter)
}
