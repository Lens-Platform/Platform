package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateBusinessAccount(t *testing.T) {
	t.Run("TestName:UpdateBusinessAccountThatDoesntExist", UpdateBusinessAccountThatDoesntExist)
	t.Run("TestName:UpdateBusinessAccountWithInvalidId", UpdateBusinessAccountWithInvalidId)
	t.Run("TestName:UpdateBusinessAccountCannotChangePassword", UpdateBusinessAccountCannotChangePassword)
	t.Run("TestName:UpdateBusinessAccount", UpdateBusinessAccount)
}

// UpdateBusinessAccount test that an account can be updated correctly
func UpdateBusinessAccount(t *testing.T) {
	ctx := context.TODO()
	var authnId uint32 = uint32(GenerateRandomId(5000, 10000))
	account := GenerateRandomizedAccount()

	password := account.Password
	// create account first
	result, err := db.CreateBusinessAccount(ctx, account, authnId)
	ExpectNoErrorOccured(t, err, result)

	var updatedCompanyName = "test_company"
	result.CompanyName = updatedCompanyName
	result.Password = password
	id := result.Id
	result.Id = 0

	// update account
	result, err = db.UpdateBusinessAccount(ctx, id, result)
	ExpectNoErrorOccured(t, err, result)
	assert.True(t, result.CompanyName == updatedCompanyName)
}

// UpdateBusinessAccountWithInvalidId test that an account can be updated correctly
func UpdateBusinessAccountWithInvalidId(t *testing.T) {
	ctx := context.TODO()
	var authnId uint32 = uint32(GenerateRandomId(5000, 10000))
	account := GenerateRandomizedAccount()
	// create account first
	result, err := db.CreateBusinessAccount(ctx, account, authnId)
	ExpectNoErrorOccured(t, err, result)

	var updatedCompanyName = "test_company"
	result.CompanyName = updatedCompanyName

	// update account
	result, err = db.UpdateBusinessAccount(ctx, 0, result)
	ExpectInvalidArgumentsError(t, err, result)
}

// UpdateBusinessAccountThatDoesntExist ensures that we cannot update a non existent account
func UpdateBusinessAccountThatDoesntExist(t *testing.T) {
	ctx := context.TODO()
	account := GenerateRandomizedAccount()
	var randomAccountId uint32 = uint32(GenerateRandomId(5000, 10000))

	result, err := db.UpdateBusinessAccount(ctx, randomAccountId, account)
	ExpectAccountDoesNotExistError(t, err, result)
}

// UpdateBusinessAccountCannotChangePassword ensures we cannot update a password as of yet
// TODO: change this when we enable this feature
func UpdateBusinessAccountCannotChangePassword(t *testing.T) {
	ctx := context.TODO()
	var authnId uint32 = uint32(GenerateRandomId(5000, 10000))
	account := GenerateRandomizedAccount()
	// create account first
	result, err := db.CreateBusinessAccount(ctx, account, authnId)
	ExpectNoErrorOccured(t, err, result)

	var newPassword = "new_password"
	result.Password = newPassword

	// update account
	result, err = db.UpdateBusinessAccount(ctx, result.Id, result)
	ExpectCannotUpdatePasswordError(t, err, result)
}
