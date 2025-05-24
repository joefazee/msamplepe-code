package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/timchuks/monieverse/internal/common"
)

func TestQueries_CreateUser(t *testing.T) {

	createRandomUser(t, "Personal")
	createRandomUser(t, "Business")
}

func createRandomUser(t *testing.T, accountType string) User {

	hashedPassword, err := common.HashPassword(common.RandomString(6))
	assert.NoError(t, err)

	arg := CreateUserParams{
		Password:    hashedPassword,
		Email:       common.RandomEmail(),
		CountryCode: faker.New().Address().CountryCode(),
		Phone:       fmt.Sprintf("%d", common.RandomInt(1000000000, 9999999999)),
		FirstName:   common.RandomOwner(),
		LastName:    common.RandomOwner(),
		MiddleName:  common.RandomOwner(),
		AccountType: accountType,
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, user)
	assert.Equal(t, arg.Email, user.Email)
	assert.Equal(t, arg.Password, user.Password)
	assert.Equal(t, arg.CountryCode, user.CountryCode)
	assert.Equal(t, arg.Phone, user.Phone)
	assert.Equal(t, arg.FirstName, user.FirstName)
	assert.Equal(t, arg.LastName, user.LastName)
	assert.Equal(t, arg.MiddleName, user.MiddleName)
	assert.Equal(t, arg.AccountType, user.AccountType)

	assert.NotZero(t, user.CreatedAt)
	assert.True(t, user.PasswordChangedAt.IsZero())

	return user

}

func TestUserOperations(t *testing.T) {

	store := SQLStore{
		db:      testDB,
		Queries: testQueries,
	}

	ctx := context.Background()

	createUserParams := CreateUserParams{
		Email:        common.RandomEmail(),
		CountryCode:  "1",
		Phone:        fmt.Sprintf("%d", common.RandomInt(1000000000, 9999999999)),
		Password:     "password",
		FirstName:    "John",
		MiddleName:   "M",
		LastName:     "Doe",
		AccountType:  "personal",
		BusinessName: "Example Business",
	}

	user, err := store.CreateUser(ctx, createUserParams)
	assert.NoError(t, err)

	otherUserEmailParams := GetOtherUserByEmailParams{
		Email: createUserParams.Email,
		ID:    uuid.New(),
	}

	_, err = store.GetOtherUserByEmail(ctx, otherUserEmailParams)
	assert.NoError(t, err)

	otherUserPhoneParams := GetOtherUserByPhoneParams{
		Phone: createUserParams.Phone,
		ID:    uuid.New(),
	}

	_, err = store.GetOtherUserByPhone(ctx, otherUserPhoneParams)
	assert.NoError(t, err)

	foundUserByEmail, err := store.GetUserByEmail(ctx, createUserParams.Email)
	assert.NoError(t, err)
	assert.Equal(t, user, foundUserByEmail)

	foundUserByPhone, err := store.GetUserByPhone(ctx, createUserParams.Phone)
	assert.NoError(t, err)
	assert.Equal(t, user, foundUserByPhone)

	updateUserParams := UpdateUserParams{
		AccountType:  NewNullString("updated"),
		Phone:        NewNullString("0987654321"),
		CountryCode:  NewNullString("2"),
		Email:        NewNullString("updated@example.com"),
		Active:       NewNullBool(true),
		Password:     NewNullString("new_password"),
		FirstName:    NewNullString("UpdatedJohn"),
		MiddleName:   NewNullString("UpdatedM"),
		LastName:     NewNullString("UpdatedDoe"),
		BusinessName: NewNullString("Updated Business"),
		Address:      NewNullString("Updated Address"),
		City:         NewNullString("Updated City"),
		State:        NewNullString("Updated State"),
		Zipcode:      NewNullString("12345"),
		ID:           user.ID,
	}

	updatedUser, err := store.UpdateUser(ctx, updateUserParams)
	assert.NoError(t, err)
	assert.NotEqual(t, user, updatedUser)
}

func TestDeactivateUser(t *testing.T) {
	user := createRandomUser(t, "Personal")

	deactivatedUser, err := testQueries.DeactivateUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, false, deactivatedUser.Active)
	//assert.NotZero(t, deactivatedUser.SuspendedAt)
	assert.Equal(t, user.ID, deactivatedUser.ID)
}

func TestUpdateUserKYCApproval(t *testing.T) {
	user := createRandomUser(t, "Personal")

	err := testQueries.UpdateUserKYCApproval(context.Background(), UpdateUserKYCApprovalParams{
		KycVerified: "approved",
		ID:          user.ID,
	})
	assert.NoError(t, err)
}
