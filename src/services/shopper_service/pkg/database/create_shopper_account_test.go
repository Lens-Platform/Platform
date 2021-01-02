package database_test

import (
	"testing"
)

func TestDb_CreateShopperAccount(t *testing.T) {
	t.Run("TestName:CreateValidShopperAccount", TestDb_CreateValidShopperAccount)
	t.Run("TestName:CreateDuplicateShopperAccount", TestDb_CreateDuplicateShopperAccount)
	t.Run("TestName:CreateShopperAccountWithInvalidConfigurations", TestDb_CreateShopperAccountWithInvalidConfigurations)
}

func TestDb_CreateShopperAccountWithInvalidConfigurations(t *testing.T){
	t.Run("TestName:CreateShopperAccountWithInvalidAuthnId", TestDb_CreateShopperAccountWithInvalidAuthnId)
	t.Run("TestName:CreateShopperAccountWithNoAccount", TestDb_CreateShopperAccountWithNoAccount)
	t.Run("TestName:CreateShopperAccountWithNoEmail", TestDb_CreateShopperAccountWithNoEmail)
	t.Run("TestName:CreateShopperAccountWithNoPassword", TestDb_CreateShopperAccountWithNoPassword)
	t.Run("TestName:CreateShopperAccountWithNoUsername", TestDb_CreateShopperAccountWithNoUsername)
	t.Run("TestName:CreateShopperAccountWithNoLastnameAndFirstname", TestDb_CreateShopperAccountWithNoLastnameAndFirstname)
}

// TestDb_CreateValidShopperAccount test an account is properly created in the backend database if a record
// doesnt already exist
func TestDb_CreateValidShopperAccount(t *testing.T) {
	testShopperAccount, authnID, ctx := DefineInitialConditions()

	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectNoError(t, err, createdAccount, authnID)
}

// TestDb_CreateDuplicateShopperAccount test an account cannot be created given a duplicate already exists
func TestDb_CreateDuplicateShopperAccount(t *testing.T) {
	testShopperAccount, authnID, ctx := DefineInitialConditions()

	// create account
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectNoError(t, err, createdAccount, authnID)

	// attempt to create duplicate account
	duplicateAccount, err := db.CreateShopperAccount(ctx, createdAccount, authnID)
	ExpectAccountAlreadyExistsError(t, err, duplicateAccount)
}

// TestDb_CreateShopperAccountWithInvalidAuthnId attempts to create an account with an invalid authentication handler service record id
func TestDb_CreateShopperAccountWithInvalidAuthnId(t *testing.T){
	testShopperAccount, authnID, ctx := DefineInitialConditions()
	// invalidate authnID
	authnID = 0

	// create account with invalid authn ID
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectInvalidArgumentsError(t, createdAccount, err)
}

// TestDb_CreateShopperAccountWithNoEmail attempts to create an account with an empty email field
func TestDb_CreateShopperAccountWithNoEmail(t *testing.T){
	testShopperAccount, authnID, ctx := DefineInitialConditions()
	// invalidate testshopper account email field
	testShopperAccount.Email = ""

	// create account with invalid shopper account
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectInvalidArgumentsError(t, createdAccount, err)
}

// TestDb_CreateShopperAccountWithNoPassword attempts to create an account with an empty password field
func TestDb_CreateShopperAccountWithNoPassword(t *testing.T){
	testShopperAccount, authnID, ctx := DefineInitialConditions()
	// invalidate testshopper account password field
	testShopperAccount.Password = ""

	// create account with invalid shopper account
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectInvalidArgumentsError(t, createdAccount, err)
}

// TestDb_CreateShopperAccountWithNoLastnameAndFirstname attempts to create an account with an empty firstname and lastname field
func TestDb_CreateShopperAccountWithNoLastnameAndFirstname(t *testing.T){
	testShopperAccount, authnID, ctx := DefineInitialConditions()
	// invalidate testshopper account lastname and firstname field
	testShopperAccount.LastName = ""
	testShopperAccount.FirstName = ""

	// create account with invalid shopper account
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectInvalidArgumentsError(t, createdAccount, err)
}

// TestDb_CreateShopperAccountWithNoUsername attempts to create an account with an empty username field
func TestDb_CreateShopperAccountWithNoUsername(t *testing.T){
	testShopperAccount, authnID, ctx := DefineInitialConditions()
	// invalidate testshopper account username field
	testShopperAccount.Username = ""

	// create account with invalid shopper account
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectInvalidArgumentsError(t, createdAccount, err)
}

// TestDb_CreateShopperAccountWithNoAccount attempts to create an account with an invalid account object
func TestDb_CreateShopperAccountWithNoAccount(t *testing.T){
	testShopperAccount, authnID, ctx := DefineInitialConditions()
	// invalidate testshopper account
	testShopperAccount = nil

	// create account with invalid shopper account
	createdAccount, err := db.CreateShopperAccount(ctx, testShopperAccount, authnID)
	ExpectInvalidArgumentsError(t, createdAccount, err)
}
