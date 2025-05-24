package db

import (
	"context"
	"database/sql"
)

func (q *Queries) GetPaginatedUserActions(ctx context.Context, filter *UserActionFilter) ([]UserAction, Metadata, error) {

	getPaginatedUserActions := `
		SELECT count(*) OVER(), id, user_id, action, message, payload, created_at FROM user_actions `
	if filter.User != nil {
		getPaginatedUserActions += ` WHERE user_id = $3`
	}
	getPaginatedUserActions += `
		ORDER BY id ASC
		LIMIT $1 OFFSET $2`

	var rows *sql.Rows
	var err error

	if filter.User != nil {
		rows, err = q.db.QueryContext(ctx, getPaginatedUserActions, filter.Limit(), filter.Offset(), filter.User.ID)
	} else {
		rows, err = q.db.QueryContext(ctx, getPaginatedUserActions, filter.Limit(), filter.Offset())
	}

	if err != nil {
		return nil, EmptyMetadata, err
	}

	defer rows.Close()

	var items []UserAction
	totalRecords := 0

	for rows.Next() {
		var i UserAction
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.UserID,
			&i.Action,
			&i.Message,
			&i.Payload,
			&i.CreatedAt,
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
