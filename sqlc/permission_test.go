package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionQueries(t *testing.T) {
	ctx := context.Background()

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	permissionParams := CreatePermissionParams{
		Name: "Test Permission",
		Code: "test-permission",
	}

	permission, err := testQueries.CreatePermission(ctx, permissionParams)
	assert.NoError(t, err)
	assert.NotNil(t, permission)

	// Test GetAll
	allPermissions, err := testQueries.GetAllPermissions(ctx)
	assert.NoError(t, err)
	assert.Greater(t, len(allPermissions), 0)

	// Test GetPermission
	fetchedPermission, err := testQueries.GetPermission(ctx, permission.ID)
	assert.NoError(t, err)
	assert.Equal(t, permission.Name, fetchedPermission.Name)
	assert.Equal(t, permission.Code, fetchedPermission.Code)

	// Test GetPermissionByCode
	fetchedPermissionByCode, err := testQueries.GetPermissionByCode(ctx, permission.Code)
	assert.NoError(t, err)
	assert.Equal(t, permission.Name, fetchedPermissionByCode.Name)
	assert.Equal(t, permission.Code, fetchedPermissionByCode.Code)

	// Create a sample user
	user := createRandomUser(t, "Personal")

	// Test GetPermissionsForUser (should be empty initially)
	userPermissions, err := testQueries.GetPermissionsForUser(ctx, user.ID)
	assert.NoError(t, err)
	assert.Empty(t, userPermissions)

	err = store.AddUserPermission(ctx, user.ID, permission.Code)
	assert.NoError(t, err)

	userPermissions, err = testQueries.GetPermissionsForUser(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(userPermissions))
	assert.Equal(t, permission.Code, userPermissions[0])

	permissionParam2 := CreatePermissionParams{
		Name: "Test Permission2",
		Code: "test-permission2",
	}

	// Test GetPermissionForUser i.e get a permission a user has
	permission2, err := testQueries.CreatePermission(ctx, permissionParam2)
	assert.NoError(t, err)
	assert.NotNil(t, permission2)

	userPermission, err := testQueries.AddPermissionForUser(ctx, AddPermissionForUserParams{UserID: user.ID, PermissionID: permission2.ID})
	assert.NoError(t, err)
	assert.NotEmpty(t, userPermission)
	assert.Equal(t, permission2.ID, userPermission.PermissionID)

	userPermission, err = testQueries.GetPermissionForUser(ctx, GetPermissionForUserParams{UserID: user.ID, PermissionID: permission2.ID})
	assert.NoError(t, err)
	assert.NotEmpty(t, userPermission)

	// Test DeletePermissionFromUser
	err = testQueries.DeletePermissionFromUser(ctx, DeletePermissionFromUserParams{UserID: user.ID, PermissionID: permission2.ID})
	assert.NoError(t, err)
	//get permission for user after deleting it
	userPermission, err = testQueries.GetPermissionForUser(ctx, GetPermissionForUserParams{UserID: user.ID, PermissionID: permission2.ID})
	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.Empty(t, userPermission)

}
