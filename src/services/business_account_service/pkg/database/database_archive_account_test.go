package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchiveBusinessAccount(t *testing.T) {
	t.Run("TestName:ArchiveBusinessAccount", ArchiveBusinessAccount)
	t.Run("TestName:ArchiveBusinessAccountWithInvalidId", ArchiveBusinessAccountWithInvalidId)
	t.Run("TestName:ArchiveBusinessAccountThatDoesntExist", ArchiveBusinessAccountThatDoesntExist)
	t.Run("TestName:ArchiveBusinessAccounts", ArchiveBusinessAccounts)
}

// ArchiveBusinessAccount test that an account can be set as inactive correctly
func ArchiveBusinessAccount(t *testing.T) {
	ctx := context.TODO()
	var authnId uint32 = uint32(GenerateRandomId(20, 100))
	account := GenerateRandomizedAccount()
	// create account first
	result, err := db.CreateBusinessAccount(ctx, account, authnId)
	ExpectNoErrorOccured(t, err, result)

	// update account
	err = db.ArchiveBusinessAccount(ctx, result.Id)
	assert.Empty(t, err)

	// get the account and ensure the account is locked
	acc := db.GetBusinessById(ctx, result.Id)
	assert.False(t, acc.IsActive)
}

// ArchiveBusinessAccountWithInvalidId test that an account can be set as inactive correctly
func ArchiveBusinessAccountWithInvalidId(t *testing.T) {
	ctx := context.TODO()
	var authnId uint32 = uint32(GenerateRandomId(20, 100))
	account := GenerateRandomizedAccount()
	// create account first
	result, err := db.CreateBusinessAccount(ctx, account, authnId)
	ExpectNoErrorOccured(t, err, result)

	// update account
	err = db.ArchiveBusinessAccount(ctx, 0)
	ExpectInvalidArgumentsError(t, err, nil)
}

// ArchiveBusinessAccountThatDoesntExist ensures that we cannot archive an account that does not exist
func ArchiveBusinessAccountThatDoesntExist(t *testing.T) {
	ctx := context.TODO()
	randomId := GenerateRandomId(500, 2000)

	// update account
	err := db.ArchiveBusinessAccount(ctx, uint32(randomId))
	ExpectAccountDoesNotExistError(t, err, nil)
}

// ArchiveBusinessAccounts archives a set of accounts
func ArchiveBusinessAccounts(t *testing.T) {
	ctx := context.TODO()
	var authnId = uint32(GenerateRandomId(20, 100))
	account := GenerateRandomizedAccount()
	// create account first
	result, _ := db.CreateBusinessAccount(ctx, account, authnId)
	var accountIds = []uint32{
		result.Id,
	}

	res, err := db.ArchiveBusinessAccounts(ctx, accountIds)
	assert.Empty(t, err)
	assert.Contains(t, res, true)
}
