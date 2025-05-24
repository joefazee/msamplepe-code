package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type CustomerFilter struct {
	Filter
	Query string `json:"query"`
	Owner *User
}

func (q *Queries) GetCustomers(ctx context.Context, filter CustomerFilter) ([]Customer, Metadata, error) {

	if filter.Owner == nil {
		return nil, EmptyMetadata, errors.New("owner is required")
	}

	var params []interface{}

	qt := `SELECT count(*) OVER() AS 
		total_records, id, name,
       		email, phone, owner_id, created_at, updated_at FROM customers WHERE owner_id = $1`
	params = append(params, filter.Owner.ID)

	searchColumns := []string{
		"name", "email", "phone",
	}

	if filter.Query != "" && len(searchColumns) > 0 {
		var searchWhere []string
		for _, column := range searchColumns {
			params = append(params, "%"+filter.Query+"%")
			searchWhere = append(searchWhere, fmt.Sprintf("LOWER(%s) LIKE $%d", column, len(params)))
		}
		qt += `AND ( ` + strings.Join(searchWhere, " OR ") + ` )`
	}

	qt += ` ORDER BY name asc`

	qt += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(params)+1, len(params)+2)
	params = append(params, filter.Limit(), filter.Offset())

	rows, err := q.db.QueryContext(ctx, qt, params...)
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()

	var customers []Customer

	totalRecords := 0
	for rows.Next() {
		var i Customer
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Name,
			&i.Email,
			&i.Phone,
			&i.OwnerID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, EmptyMetadata, err
		}
		customers = append(customers, i)
	}
	if err := rows.Close(); err != nil {
		return nil, EmptyMetadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, EmptyMetadata, err
	}

	metadata := CalculateMetadata(totalRecords, filter.Page, filter.Limit())
	return customers, metadata, nil
}
