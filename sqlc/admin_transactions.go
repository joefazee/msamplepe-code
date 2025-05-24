package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetPaginatedTransactionRow struct {
	ID                uuid.UUID       `json:"id"`
	UserID            uuid.UUID       `json:"user_id"`
	Amount            decimal.Decimal `json:"amount"`
	Customer          string          `json:"customer"`
	Status            string          `json:"status"`
	Rate              decimal.Decimal `json:"rate"`
	CreatedAt         time.Time       `json:"created_at"`
	PaymentMethod     string          `json:"payment_method"`
	Source            string          `json:"source"`
	Action            string          `json:"action"`
	Currency          string          `json:"currency"`
	Payload           json.RawMessage `json:"payload"`
	IsInvoiceUploaded bool            `json:"is_invoice_uploaded"`
	Type              string          `json:"type"`
}

func (store *SQLStore) AdminGetPaginatedTransactionList(ctx context.Context, filter *TransactionFilter) ([]GetPaginatedTransactionRow, Metadata, error) {

	var transactions []GetPaginatedTransactionRow

	query := buildPaginatedQueryForAdminGetTransactions(filter)
	rows, err := store.db.QueryContext(ctx, query, filter.params...)
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()

	totalRecords := 0
	for rows.Next() {
		var i GetPaginatedTransactionRow
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Amount,
			&i.Status,
			&i.CreatedAt,
			&i.UserID,
			&i.Rate,
			&i.PaymentMethod,
			&i.Source,
			&i.Action,
			&i.Currency,
			&i.Payload,
			&i.Customer,
			&i.IsInvoiceUploaded,
			&i.Type,
		); err != nil {
			return nil, EmptyMetadata, err
		}
		transactions = append(transactions, i)
	}
	if err := rows.Close(); err != nil {
		return nil, EmptyMetadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, EmptyMetadata, err
	}

	metadata := CalculateMetadata(totalRecords, filter.Page, filter.Limit())
	return transactions, metadata, nil
}

func buildPaginatedQueryForAdminGetTransactions(filter *TransactionFilter) string {
	columns := []string{
		"t.id",
		"t.amount",
		"t.status",
		"t.created_at",
		"t.user_id",
		"t.rate",
		"t.payment_method",
		"t.source",
		"t.action",
		"c.code as currency",
		"t.payload",
		"CONCAT(u.first_name, ' ', u.last_name) AS customer",
		"t.is_invoice_uploaded",
		"t.type",
	}

	columns = append(columns, filter.additionalSelectFields...)

	searchColumns := []string{
		"u.first_name",
		"u.last_name",
		"u.email",
		"u.phone",
		"t.type",
	}

	qt := fmt.Sprintf(`SELECT count(*) OVER() AS total_records, %s FROM transactions t`, strings.Join(columns, ", "))
	qt += " LEFT JOIN users u ON u.id = t.user_id"
	qt += " LEFT JOIN currencies c ON c.id = t.currency_id"

	if len(filter.additionalLeftJoin) > 0 {
		qt += strings.Join(filter.additionalLeftJoin, " ")
	}

	if filter.Status != "" {
		filter.WhereConditions = append(filter.WhereConditions, fmt.Sprintf("t.status = $%d", len(filter.params)+1))
		filter.params = append(filter.params, filter.Status)
	}

	if len(filter.WhereConditions) > 0 {
		qt += " WHERE " + strings.Join(filter.WhereConditions, " AND ")
	}

	if filter.DateBetween != "" {
		qt += buildDateRangeSQL(filter.DateBetween, filter)
	}

	if filter.Query != "" && len(searchColumns) > 0 {
		var searchWhere []string

		for _, field := range searchColumns {
			searchWhere = append(searchWhere, fmt.Sprintf("%s = $%d", field, len(filter.params)+1))
			filter.params = append(filter.params, filter.Query)
		}

		qt += " AND (" + strings.Join(searchWhere, " OR ") + ")"
	}

	if len(filter.groupBy) > 0 {
		qt += filter.groupBy
	}

	if len(filter.OrderBy) > 0 {
		qt += " ORDER BY " + strings.Join(filter.OrderBy, ", ")
	}

	qt += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(filter.params)+1, len(filter.params)+2)
	filter.params = append(filter.params, filter.Limit(), filter.Offset())

	return qt
}
