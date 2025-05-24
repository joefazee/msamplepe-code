package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionSearchParam struct {
	CustomerFirstName string `json:"customer_first_name,omitempty"`
	CustomerLastName  string `json:"customer_last_name,omitempty"`
	CustomerEmail     string `json:"customer_email,omitempty"`
}

type Join struct {
	Table       string
	JoinType    string
	OnCondition string
}
type QueryOptions struct {
	Joins   []Join
	OrderBy []string
	Where   map[string]any
	Search  map[string]any
	Params  []interface{}
}

type TransactionFilter struct {
	Filter
	Query                  string
	OrderBy                []string
	Status                 string
	DateBetween            string // date format: 2020-01-01,2020-01-31
	params                 []interface{}
	WhereConditions        []string
	User                   *User
	additionalSelectFields []string
	groupBy                string
	additionalLeftJoin     []string
	Type                   string
}

type SwapRequestTransactionRow struct {
	ID                uuid.UUID       `json:"id"`
	Amount            decimal.Decimal `json:"amount"`
	Customer          string          `json:"customer"`
	Status            string          `json:"status"`
	Rate              decimal.Decimal `json:"rate"`
	CreatedAt         time.Time       `json:"created_at"`
	UserID            uuid.UUID       `json:"user_id"`
	PaymentMethod     string          `json:"payment_method"`
	FeesAmount        decimal.Decimal `json:"fees_amount"`
	Source            string          `json:"source"`
	Action            string          `json:"action"`
	Currency          string          `json:"currency"`
	Payload           json.RawMessage `json:"payload"`
	IsInvoiceUploaded bool            `json:"is_invoice_uploaded"`
	TotalSettled      decimal.Decimal `json:"total_settled"`
	Gateway           string          `json:"gateway"`
	TrackingNumber    string          `json:"tracking_number"`
	CompanyName       string          `json:"company_name"`
	Tag               string          `json:"tag"`
	Type              string          `json:"type"`
}

type TransactionRow struct {
	ID                uuid.UUID       `json:"id"`
	Amount            decimal.Decimal `json:"amount"`
	Customer          string          `json:"customer"`
	Status            string          `json:"status"`
	Rate              decimal.Decimal `json:"rate"`
	CreatedAt         time.Time       `json:"created_at"`
	UserID            uuid.UUID       `json:"user_id"`
	PaymentMethod     string          `json:"payment_method"`
	FeesAmount        decimal.Decimal `json:"fees_amount"`
	Source            string          `json:"source"`
	Action            string          `json:"action"`
	Currency          string          `json:"currency"`
	Payload           json.RawMessage `json:"payload"`
	IsInvoiceUploaded bool            `json:"is_invoice_uploaded"`
	Gateway           string          `json:"gateway"`
	TrackingNumber    string          `json:"tracking_number"`
	CompanyName       string          `json:"company_name"`
	Tag               string          `json:"tag"`
	Type              string          `json:"type"`
}

func (q *Queries) GetPaginatedSwapRequests(ctx context.Context, filter *TransactionFilter) ([]SwapRequestTransactionRow, Metadata, error) {

	filter.additionalSelectFields = []string{
		"CAST(COALESCE(SUM(s.amount), 0) AS numeric) AS total_settled",
	}

	filter.groupBy = " GROUP BY t.id, t.amount, t.status, t.created_at, t.user_id, t.rate, t.payment_method, t.source, t.action, c.code, t.payload, u.first_name, u.last_name"

	filter.additionalLeftJoin = []string{
		" LEFT JOIN settlements s ON t.id = s.transaction_id AND s.status = 'settled'",
	}
	var transactions []SwapRequestTransactionRow

	filter.WhereConditions = []string{
		"t.requires_settlement = true",
		"t.source = 'wallet'",
	}

	query := buildPaginatedQuery(filter)
	rows, err := q.db.QueryContext(ctx, query, filter.params...)
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()

	totalRecords := 0
	for rows.Next() {
		var i SwapRequestTransactionRow
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Amount,
			&i.Status,
			&i.Tag,
			&i.CreatedAt,
			&i.UserID,
			&i.Rate,
			&i.PaymentMethod,
			&i.FeesAmount,
			&i.Source,
			&i.Action,
			&i.Currency,
			&i.Payload,
			&i.Customer,
			&i.IsInvoiceUploaded,
			&i.Gateway,
			&i.TrackingNumber,
			&i.CompanyName,
			&i.Type,
			&i.TotalSettled,
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

func (store *SQLStore) GetPaginatedTransactions(ctx context.Context, filter *TransactionFilter) ([]TransactionRow, Metadata, error) {
	var transactions []TransactionRow

	if filter.User != nil {
		filter.WhereConditions = append(filter.WhereConditions, fmt.Sprintf("t.user_id = $%d", len(filter.params)+1))
		filter.params = append(filter.params, filter.User.ID)
	}

	query := buildPaginatedQuery(filter)
	rows, err := store.db.QueryContext(ctx, query, filter.params...)
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()

	totalRecords := 0
	for rows.Next() {
		var i TransactionRow
		var tag sql.NullString
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Amount,
			&i.Status,
			&tag,
			&i.CreatedAt,
			&i.UserID,
			&i.Rate,
			&i.PaymentMethod,
			&i.FeesAmount,
			&i.Source,
			&i.Action,
			&i.Currency,
			&i.Payload,
			&i.Customer,
			&i.IsInvoiceUploaded,
			&i.Gateway,
			&i.TrackingNumber,
			&i.CompanyName,
			&i.Type,
		); err != nil {
			return nil, EmptyMetadata, err
		}

		i.Tag = NullStringToString(tag)
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

// buildPaginatedQuery returns a paginated SELECT with dynamic WHERE / AND / OR
// parts.  JSON-payload search is executed **only** for rows whose
// t.action = 'ext-transfer', and every JSON value is cast to text before
// using ILIKE so Postgres doesn’t raise “jsonb ~~* unknown”.
func buildPaginatedQuery(filter *TransactionFilter) string {

	columns := []string{
		"t.id",
		"t.amount",
		"t.status",
		"t.tag",
		"t.created_at",
		"t.user_id",
		"t.rate",
		"t.payment_method",
		"t.fees_amount",
		"t.source",
		"t.action",
		"c.code AS currency",
		"t.payload",
		"CONCAT(u.first_name, ' ', u.last_name) AS customer",
		"t.is_invoice_uploaded",
		"t.gateway",
		"t.tracking_number",
		"t.company_name",
		"t.type",
	}
	columns = append(columns, filter.additionalSelectFields...)

	searchColumns := []string{
		"u.first_name",
		"u.last_name",
		"u.email",
		"u.phone",
	}
	jsonSearchFields := []string{
		"t.payload->'recipient'->'data'->>'account_name'",
		"t.payload->'recipient'->'data'->>'email'",
		"t.payload->'recipient'->'data'->>'bank_name'",
		"t.payload->'recipient'->'data'->>'recipient_address'",
		"t.payload->'recipient'->'data'->>'business_name'",
		"t.payload->'recipient'->'data'->>'account_number'",
	}
	searchColumns = append(searchColumns, jsonSearchFields...)

	qt := fmt.Sprintf(
		`SELECT count(*) OVER() AS total_records, %s FROM transactions t`,
		strings.Join(columns, ", "),
	)
	qt += " LEFT JOIN users u ON u.id = t.user_id"
	qt += " LEFT JOIN currencies c ON c.id = t.currency_id"
	if len(filter.additionalLeftJoin) > 0 {
		qt += " " + strings.Join(filter.additionalLeftJoin, " ")
	}

	if filter.Status != "" {
		filter.WhereConditions = append(
			filter.WhereConditions,
			fmt.Sprintf("t.status = $%d", len(filter.params)+1),
		)
		filter.params = append(filter.params, filter.Status)
	}
	if filter.Type != "" {
		filter.WhereConditions = append(
			filter.WhereConditions,
			fmt.Sprintf("t.type = $%d", len(filter.params)+1),
		)
		filter.params = append(filter.params, filter.Type)
	}

	hasWhere := false
	if len(filter.WhereConditions) > 0 {
		qt += " WHERE " + strings.Join(filter.WhereConditions, " AND ")
		hasWhere = true
	}

	if filter.DateBetween != "" {
		dateExpr := buildDateRangeSQL(filter.DateBetween, filter)
		if dateExpr != "" {
			if !hasWhere {
				qt += " WHERE "
				hasWhere = true
			} else {
				qt += " AND "
			}
			qt += dateExpr
		}
	}

	if filter.Query != "" {
		var parts []string

		for _, field := range searchColumns {
			placeholder := fmt.Sprintf("$%d", len(filter.params)+1)

			if strings.Contains(field, "payload") { // JSON path
				parts = append(parts,
					fmt.Sprintf(
						"(t.action = 'ext-transfer' AND (%s)::text ILIKE %s)",
						field, placeholder,
					),
				)
				filter.params = append(filter.params, "%"+filter.Query+"%")
			} else {
				parts = append(parts, fmt.Sprintf("%s = %s", field, placeholder))
				filter.params = append(filter.params, filter.Query)
			}
		}

		placeholder := fmt.Sprintf("$%d", len(filter.params)+1)
		parts = append(parts,
			fmt.Sprintf(
				"(t.action = 'ext-transfer' AND "+
					"(t.payload->'recipient'->'data')::text ILIKE %s)",
				placeholder,
			),
		)
		filter.params = append(filter.params, "%"+filter.Query+"%")

		if !hasWhere {
			qt += " WHERE "
			hasWhere = true
		} else {
			qt += " AND "
		}
		qt += "(" + strings.Join(parts, " OR ") + ")"
	}

	if len(filter.groupBy) > 0 {
		qt += filter.groupBy
	}
	if len(filter.OrderBy) > 0 {
		qt += " ORDER BY " + strings.Join(filter.OrderBy, ", ")
	}

	qt += fmt.Sprintf(" LIMIT $%d OFFSET $%d",
		len(filter.params)+1, len(filter.params)+2)
	filter.params = append(filter.params, filter.Limit(), filter.Offset())

	return qt
}

func buildDateRangeSQL(between string, filter *TransactionFilter) string {
	parts := strings.Split(between, ",")
	var fromDate string
	var toDate string

	if len(parts) == 2 {
		fromDate = parts[0]
		toDate = parts[1]
	} else {
		fromDate = parts[0]
		toDate = parts[0]
	}

	q := fmt.Sprintf(" AND t.created_at BETWEEN $%d AND $%d", len(filter.params)+1, len(filter.params)+2)

	filter.params = append(filter.params, fromDate, toDate)
	return q
}
