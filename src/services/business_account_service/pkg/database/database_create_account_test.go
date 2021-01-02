package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBusinessCreateAccount(t *testing.T) {
	t.Run("TestName:CreateBusinessAccount", CreateAccountNonExistentAccount)
	t.Run("TestName:CreateDuplicateBusinessAccounts", CreateDuplicateBusinessAccounts)
	t.Run("TestName:CreateBusinessAccountWithFaultyAuthnId", CreateBusinessAccountWithFaultyAuthnId)
	t.Run("TestName:CreateBusinessAccountWithNoCompanyName", CreateBusinessAccountWithNoCompanyName)
	t.Run("TestName:CreateBusinessAccountWithNoEmail", CreateBusinessAccountWithNoEmail)
	t.Run("TestName:CreateBusinessAccountWithNoPassword", CreateAccountNonExistentAccount)
}

// CreateAccountNonExistentAccount test an account is properly created in the backend database if a record
// doesnt already exist
func CreateAccountNonExistentAccount(t *testing.T) {
	account := GenerateRandomizedAccount()

	createdAccount, err := db.CreateBusinessAccount(context.Background(), account, 1)
	assert.Empty(t, err)
	assert.NotEmpty(t, createdAccount, "user record cannot be empty")
	// assert the user record returned is actually active
	assert.Equal(t, createdAccount.IsActive, true, "user record should be activated")
}

// CreateDuplicateBusinessAccounts tests that duplicate accounts can have the expected error return type
func CreateDuplicateBusinessAccounts(t *testing.T) {
	account := GenerateRandomizedAccount()

	var authnId uint32 = 500
	// create account twice
	createdAccount, err := db.CreateBusinessAccount(context.Background(), account, authnId)
	assert.Empty(t, err)

	createdAccount, err = db.CreateBusinessAccount(context.Background(), createdAccount, authnId)
	// an error should occur
	ExpectAccountAlreadyExistError(t, err, createdAccount)
}

// CreateBusinessAccountWithFaultyAuthnId tests the proper errors are returned for accounts with faulty authnn Ids
func CreateBusinessAccountWithFaultyAuthnId(t *testing.T) {
	account := GenerateRandomizedAccount()

	var authnId uint32 = 0
	// create account with faulty Id and ensure the proper expected error is returned
	createdAccount, err := db.CreateBusinessAccount(context.Background(), account, authnId)
	ExpectInvalidArgumentsError(t, err, createdAccount)
}

// CreateBusinessAccountWithNoCompanyName tests the proper errors are returned for accounts with no company name
func CreateBusinessAccountWithNoCompanyName(t *testing.T) {
	account := GenerateRandomizedAccount()
	// remove company name
	account.CompanyName = ""

	var authnId uint32 = 501
	createdAccount, err := db.CreateBusinessAccount(context.Background(), account, authnId)
	ExpectInvalidArgumentsError(t, err, createdAccount)
}

// CreateBusinessAccountWithNoEmail tests the proper errors are returned for accounts with no email
func CreateBusinessAccountWithNoEmail(t *testing.T) {
	account := GenerateRandomizedAccount()
	// remove company name
	account.Email = ""

	var authnId uint32 = 502
	createdAccount, err := db.CreateBusinessAccount(context.Background(), account, authnId)
	ExpectInvalidArgumentsError(t, err, createdAccount)
}

// CreateBusinessAccountWithNoPassword tests the proper errors are returned for accounts with no password
func CreateBusinessAccountWithNoPassword(t *testing.T) {
	account := GenerateRandomizedAccount()
	// remove company name
	account.Password = ""

	var authnId uint32 = 503
	createdAccount, err := db.CreateBusinessAccount(context.Background(), account, authnId)
	ExpectInvalidArgumentsError(t, err, createdAccount)
}
