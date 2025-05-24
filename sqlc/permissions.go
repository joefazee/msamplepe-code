package db

import "context"

func (q *Queries) HasPermission(ctx context.Context, user User, perm string) bool {
	permissions, err := q.GetPermissionsForUser(ctx, user.ID)
	if err != nil {
		return false
	}
	if !PermissionsInclude(permissions, perm) {
		return false
	}

	return true
}
