package db

import (
	"context"
	"fmt"

	"github.com/lib/pq"
)

type GetKYCRequirementsForUserParams struct {
	UserID      string
	CountryCode string
	AccountType string
	Statuses    []string
	Limit       int
	Offset      int
}

const getKYCRequirementsForUser = `-- name: GetKYCRequirementsForUser :many
SELECT 
    id, 
    title, 
    description, 
    payload, 
    target, 
    target_id, 
    deadline, 
    status, 
    created_at 
FROM 
    kyc_requirements 
WHERE 
    status = ANY($4)
    AND (
        target = 'all'
        OR (target = 'user' AND target_id = $1)
        OR (target = 'country' AND target_id = $2)
        OR (target = 'account_type' AND target_id = $3)
    )
ORDER BY 
    deadline ASC NULLS LAST
LIMIT $5 OFFSET $6;
`

func (store *SQLStore) GetKYCRequirementsForUser(ctx context.Context, arg GetKYCRequirementsForUserParams) ([]KycRequirement, error) {
	rows, err := store.db.QueryContext(ctx, getKYCRequirementsForUser,
		arg.UserID,
		arg.CountryCode,
		arg.AccountType,
		pq.Array(arg.Statuses),
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("GetKYCRequirementsForUser: %w", err)
	}
	defer rows.Close()

	var items []KycRequirement
	for rows.Next() {
		var i KycRequirement
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Payload,
			&i.Target,
			&i.TargetID,
			&i.Deadline,
			&i.Status,
			&i.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetKYCRequirementsForUser: row.Scan: %w", err)
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("GetKYCRequirementsForUser: rows.Close: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetKYCRequirementsForUser: rows.Err: %w", err)
	}
	return items, nil
}
