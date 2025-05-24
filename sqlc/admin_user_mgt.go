package db

import (
	"context"

	"time"

	"github.com/google/uuid"
)

const getPaginatedUsersList = `-- name: GetPaginatedUsersList :many
SELECT 
	count(*) OVER() AS total_count,
	id, email, country_code,phone,password,first_name,middle_name,last_name,account_type,business_name,active,kyc_verified,created_at,suspended_at
FROM 
	users 
WHERE
	email ILIKE $1 || '%' AND first_name ILIKE $2 || '%' AND last_name ILIKE $3 || '%' AND business_name ILIKE $4 || '%' 
	AND account_type ILIKE $5 || '%' AND account_number ILIKE $6 || '%'
ORDER BY 
	last_name ASC 
LIMIT $7 OFFSET $8
`

type GetPaginatedUserRow struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	CountryCode  string    `json:"country_code"`
	Phone        string    `json:"phone"`
	Password     string    `json:"-"`
	FirstName    string    `json:"first_name"`
	MiddleName   string    `json:"middle_name"`
	LastName     string    `json:"last_name"`
	Active       bool      `json:"active"`
	AccountType  string    `json:"account_type"`
	BusinessName string    `json:"business_name"`
	KycVerified  string    `json:"kyc_verified"`
	CreatedAt    time.Time `json:"created_at"`
	SuspendedAt  any       `json:"suspended_at"`
}

func (q *Queries) GetPaginatedUsersList(ctx context.Context, filter UserListFilter) ([]GetPaginatedUserRow, Metadata, error) {

	rows, err := q.db.QueryContext(ctx, getPaginatedUsersList,
		filter.Search.Email, filter.Search.FirstName, filter.Search.LastName,
		filter.Search.BusinessName, filter.Search.AccountType, filter.Search.AccountNumber, filter.Limit(), filter.Offset())
	if err != nil {
		return nil, EmptyMetadata, err
	}
	defer rows.Close()
	var items []GetPaginatedUserRow
	totalRecords := 0
	for rows.Next() {
		var i GetPaginatedUserRow
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.Email,
			&i.CountryCode,
			&i.Phone,
			&i.Password,
			&i.FirstName,
			&i.MiddleName,
			&i.LastName,
			&i.AccountType,
			&i.BusinessName,
			&i.Active,
			&i.KycVerified,
			&i.CreatedAt,
			&i.SuspendedAt,
		); err != nil {
			return nil, EmptyMetadata, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, EmptyMetadata, err
	}
	if err := rows.Err(); err != nil {
		return nil, EmptyMetadata, err
	}

	metadata := CalculateMetadata(totalRecords, filter.Page, filter.Limit())
	return items, metadata, nil
}
