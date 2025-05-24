package db

import (
	"context"
	"fmt"
	"strings"
)

func (q *Queries) PaginatedBusinesses(ctx context.Context, filter BusinessListFilter) ([]Business, error) {
	searchColumns := []string{
		"name", "registration_number", "business_nature", "business_category",
		"address1", "address2", "city", "post_code", "state", "country",
		"website", "product_description", "phone", "email", "contact_name",
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if len(filter.Search) > 0 {
		var searchConditions []string
		for _, col := range searchColumns {
			searchConditions = append(searchConditions, fmt.Sprintf("%s ILIKE $%d", col, argIdx))
			args = append(args, "%"+filter.Search+"%")
			argIdx++
		}
		conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("approval_status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.CreatedAtFrom != "" {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, filter.CreatedAtFrom)
		argIdx++
	}
	if filter.CreatedAtTo != "" {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, filter.CreatedAtTo)
		argIdx++
	}

	query := `SELECT id, name, registration_number, business_nature, business_category, address1, address2, city, post_code, state, country, website, product_description, phone, email, contact_name, created_at, updated_at, created_by, approval_status, approval_status_reason, approval_status_updated_by, approval_status_updated_at FROM businesses`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY name ASC LIMIT $" + fmt.Sprintf("%d", argIdx) + " OFFSET $" + fmt.Sprintf("%d", argIdx+1)
	args = append(args, filter.Limit(), filter.Offset())

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Business
	for rows.Next() {
		var i Business
		if err := rows.Scan(
			&i.ID, &i.Name, &i.RegistrationNumber, &i.BusinessNature, &i.BusinessCategory,
			&i.Address1, &i.Address2, &i.City, &i.PostCode, &i.State, &i.Country, &i.Website,
			&i.ProductDescription, &i.Phone, &i.Email, &i.ContactName, &i.CreatedAt, &i.UpdatedAt, &i.CreatedBy,
			&i.ApprovalStatus, &i.ApprovalStatusReason, &i.ApprovalStatusUpdatedBy, &i.ApprovalStatusUpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}
