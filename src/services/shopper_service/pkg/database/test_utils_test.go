package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
)

// ExpectInvalidArgumentsError expects an error of type invalid arguments to have occured from the operation
func ExpectInvalidArgumentsError(t *testing.T, createdAccount *model.ShopperAccount, err error) {
	assert.Empty(t, createdAccount)
	assert.Error(t, err)
	assert.Equal(t, err, svcErrors.ErrInvalidInputArguments)
}

// ExpectAccountAlreadyExistsError expects an account already exists error is present
func ExpectAccountAlreadyExistsError(t *testing.T, err error, account *model.ShopperAccount) {
	assert.NotEmpty(t, err)
	assert.Empty(t, account, "user record must be empty")
	assert.Equal(t, err, svcErrors.ErrAccountAlreadyExist)
}

// ExpectCannotUpdatePasswordError expects an acount password cannot be updated error
func ExpectCannotUpdatePasswordError(t *testing.T, err error, account *model.ShopperAccount) {
	assert.NotEmpty(t, err)
	assert.Empty(t, account, "user record must be empty")
	assert.Equal(t, err, svcErrors.ErrAccountDoesNotExist)
}

// ExpectAccountDoesNotExistsError expects an account does not exists error is present
func ExpectAccountDoesNotExistsError(t *testing.T, err error, account *model.ShopperAccount) {
	assert.NotEmpty(t, err)
	assert.Empty(t, account, "user record must be empty")
	assert.Equal(t, err, svcErrors.ErrAccountDoesNotExist)
}

// ExpectNoError performs a few checks ensuring no errors occured during the operation
func ExpectNoError(t *testing.T, err error, createdAccount *model.ShopperAccount, authnID uint32) {
	assert.Empty(t, err)
	assert.NotEmpty(t, createdAccount, "user record cannot be empty")
	// assert the user record returned is actually active
	assert.Equal(t, createdAccount.IsActive, true, "user record should be activated")
	assert.NotNil(t, createdAccount.Causes)
	assert.NotNil(t, createdAccount.SubscribedTopics)
	assert.NotNil(t, createdAccount.Tags)
	assert.NotNil(t, createdAccount.Addresses)
	assert.NotNil(t, createdAccount.CreditCard)
	assert.Equal(t, createdAccount.AuthnId, authnID)
}
