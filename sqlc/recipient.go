package db

import (
	"context"
	"fmt"
	"strings"
)

type RecipientFilter struct {
	Filter
	Query string `json:"query" form:"query"`
	User  *User  `json:"-"`
}

func (store *SQLStore) GetPaginatedRecipients(ctx context.Context, filter *RecipientFilter) ([]Recipient, Metadata, error) {

	qt, params := buildRecipientsQuery(filter)

	rows, err := store.db.QueryContext(ctx, qt, params...)
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()

	var recipients []Recipient

	totalRecords := 0
	for rows.Next() {
		var i Recipient
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.UserID,
			&i.Scheme,
			&i.Currency,
			&i.Data,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, EmptyMetadata, err
		}
		recipients = append(recipients, i)
	}
	if err := rows.Close(); err != nil {
		return nil, EmptyMetadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, EmptyMetadata, err
	}

	metadata := CalculateMetadata(totalRecords, filter.Page, filter.Limit())
	return recipients, metadata, nil

}

func buildRecipientsQuery(filter *RecipientFilter) (string, []interface{}) {
	var params []interface{}

	columns := []string{
		"r.id",
		"r.user_id",
		"r.scheme",
		"r.currency",
		"r.data",
		"r.created_at",
		"r.updated_at",
	}

	searchColumns := []string{
		"r.data::text",
		"r.currency::text",
	}

	qt := `SELECT count(*) OVER() AS total_records, ` + strings.Join(columns, ", ") + ` FROM recipients r`

	if filter.User != nil {
		params = append(params, filter.User.ID)
		qt += ` WHERE r.user_id = $` + fmt.Sprintf("%d", len(params))
	}

	if filter.Query != "" && len(searchColumns) > 0 {
		var searchWhere []string
		for _, column := range searchColumns {
			params = append(params, "%"+filter.Query+"%")
			searchWhere = append(searchWhere, fmt.Sprintf("LOWER(%s) LIKE $%d", column, len(params)))
		}

		if filter.User == nil {
			qt += ` WHERE `
		} else {
			qt += ` AND `
		}
		qt += `( ` + strings.Join(searchWhere, " OR ") + ` )`
	}

	qt += ` ORDER BY r.created_at DESC`

	qt += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(params)+1, len(params)+2)
	params = append(params, filter.Limit(), filter.Offset())

	return qt, params

}
