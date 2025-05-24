package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/timchuks/monieverse/internal/settings"
)

func (store *SQLStore) GetUserSettings(ctx context.Context, userID uuid.UUID) (settings.UserSettings, error) {

	var userSettings settings.UserSettings

	mapConfigs := make(map[string]string)

	rows, err := store.getUserSettings(ctx, userID)
	if err != nil {
		return userSettings, err
	}

	for _, row := range rows {
		mapConfigs[row.Key] = row.Value
	}

	err = store.mapper.Map(mapConfigs, &userSettings)
	if err != nil {
		return userSettings, err
	}

	return userSettings, nil

}

const getUserSettings = `-- name: GetUserSettings :many
SELECT id, user_id, key, value, created_at, updated_at FROM user_settings WHERE user_id = $1
`

func (q *Queries) getUserSettings(ctx context.Context, userID uuid.UUID) ([]UserSetting, error) {
	rows, err := q.db.QueryContext(ctx, getUserSettings, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UserSetting{}
	for rows.Next() {
		var i UserSetting
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Key,
			&i.Value,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
