package database_test

import (
	"context"
	"testing"

	core_database "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-database"
	"github.com/stretchr/testify/assert"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
)

func TestDb_UpdateShopperAccount(t *testing.T){
	t.Run("TestName:UpdateExistingShopperAccount", TestDB_UpdateExistingAccount)
	t.Run("TestName:UpdateExistentShopperAccount", TestDB_UpdateNonExistentAccount)
	t.Run("TestName:UpdateExistentShopperAccountPasswordField", TestDB_UpdateExistentShopperAccountPasswordField)
	t.Run("TestName:UpdateShopperAccountWithInvalidConfigurations", TestDb_UpdateShopperAccountWithInvalidConfigurations)
}

// TestDB_UpdateExistingAccount attempts to properly update an existing account
func TestDB_UpdateExistingAccount(t *testing.T){
	authnID, ctx, createdAccount, err := CreateAccountInitialCondition(t)
	if createdAccount == nil {
		t.Fatal()
	}

	// attempt to update existing account
	randStr := core_database.GenerateRandomString(150)
	createdAccount.Email = "test_updated_email" + randStr
	updatedAccount, err := db.UpdateShopperAccount(ctx, createdAccount.Id, createdAccount)
	ExpectNoError(t, err, updatedAccount, authnID)
	assert.Equal(t, updatedAccount.Email, createdAccount.Email)
}

// TestDB_UpdateNonExistentAccount attempts to update an account that does not exist
func TestDB_UpdateNonExistentAccount(t *testing.T){
	testShopperAccount, _, ctx := DefineInitialConditions()

	randomId := GenerateRandomNumber()
	updatedAccount, err := db.UpdateShopperAccount(ctx, randomId, testShopperAccount)
	ExpectAccountDoesNotExistsError(t, err, updatedAccount)
}

// TestDB_UpdateExistentShopperAccountPasswordField attempts to update an account password field
func TestDB_UpdateExistentShopperAccountPasswordField(t *testing.T){
	_, ctx, createdAccount, err := CreateAccountInitialCondition(t)
	if createdAccount == nil {
		t.Fatal()
	}

	randStr := core_database.GenerateRandomString(150)
	createdAccount.Password = "test_updated_password" + randStr
	updatedAccount, err := db.UpdateShopperAccount(ctx, createdAccount.Id, createdAccount)
	ExpectCannotUpdatePasswordError(t, err, updatedAccount)
}

// TestDb_UpdateShopperAccountWithInvalidConfigurations attempts to update an account with misconfigured input
func TestDb_UpdateShopperAccountWithInvalidConfigurations(t *testing.T){
	_, ctx, createdAccount, err := CreateAccountInitialCondition(t)
	if createdAccount == nil {
		t.Fatal()
	}

	// define misconfigurations
	createdAccount.Password = ""
	createdAccount.Username = ""
	createdAccount.FirstName = ""
	createdAccount.LastName = ""
	createdAccount.Email = ""
	createdAccount.Id = 0

	updatedAccount, err := db.UpdateShopperAccount(ctx, createdAccount.Id, createdAccount)
	ExpectInvalidArgumentsError(t, updatedAccount, err)

	updatedAccount, err = db.UpdateShopperAccount(ctx, createdAccount.Id, nil)
	ExpectInvalidArgumentsError(t, updatedAccount, err)
}

// CreateAccountInitialCondition creates an account as an initial condition
func CreateAccountInitialCondition(t *testing.T) (uint32, context.Context, *model.ShopperAccount, error) {
	testShopperAccount, authnID, ctx := DefineInitialConditions()
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectNoError(t, err, createdAccount, authnID)
	return authnID, ctx, createdAccount, err
}
