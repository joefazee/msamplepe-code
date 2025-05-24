package db

import (
	"context"

	"github.com/google/uuid"
)

func (store *SQLStore) GetOneTokenForUser(ctx context.Context, userID uuid.UUID, scope string) (Token, error) {
	query := `
	SELECT
	    t1.*
	FROM
	    tokens t1
	INNER JOIN
	    (
	        SELECT
	            user_id,
	            MAX(created_at) AS latest_created_at
	        FROM
	            tokens
	        WHERE
	            expiry > NOW()
	            AND user_id = $1
	            AND scope = $2
	        GROUP BY
	            user_id
	    ) t2
	ON
	    t1.user_id = t2.user_id
	    AND t1.created_at = t2.latest_created_at
	WHERE
	    t1.expiry > NOW()
	    AND t1.scope = $2
	    AND t1.user_id = $1;
	`

	var token Token

	err := store.db.QueryRowContext(ctx, query, userID, scope).Scan(
		&token.Hash,
		&token.UserID,
		&token.Expiry,
		&token.Scope,
		&token.CreatedAt,
		&token.Provider,
	)

	if err != nil {
		return Token{}, err
	}

	return token, nil
}
