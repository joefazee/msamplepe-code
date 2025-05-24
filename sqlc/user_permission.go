package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PermissionsInclude check if a code exists in permissions array
func PermissionsInclude(perms []string, code string) bool {
	for i := range perms {
		if code == perms[i] {
			return true
		}
	}
	return false
}

func (store *SQLStore) AddUserPermission(ctx context.Context, userID uuid.UUID, codes ...string) error {

	query := `
	INSERT INTO users_permissions
	SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)
	`

	_, err := store.db.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
