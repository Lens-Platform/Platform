package graphql_api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	middleware "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-middleware"
	"github.com/stretchr/testify/assert"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/generated"
)

var (
	CreateBusinessAccountOperationName = "CreateBusinessAccount"
)

func TestE2ECreateAccount(t *testing.T) {
	t.Run("TestName:E2E_CreateAccount", CreateAccount)
	// case where account exists but is active
	t.Run("TestName:E2E_CreateExistentActiveAccount", CreateExistentActiveAccount)
	// case where account exists but is inactive
	t.Run("TestName:E2E_CreateExistentInactiveAccount", CreateExistentInactiveAccount)
}

// CreateAccount tests the create account scenario where we create a previously non-existent account record
func CreateAccount(t *testing.T) {
	/*
			Since this is an integration test, we generate account data.
			1. we first initiate a call to the authentication handler service to create a record in the authentication service from it's perspective
		and return to us an "authn" id as well as a json web token.
			2. we pass this token into a context which we wrap in our custom middleware function.
		This ensures the context gets propagated to the graphql mutation handler
			3. we perform our request with and expect no error to occur
	*/
	randStr := graphql_api.GenerateRandomString(40)
	fakeEmail := randStr + "@gmail.com"
	fakeCompanyName := randStr
	fakePassword := randStr + "pwd"

	token, _ := graphql_api.CreateAccountInAuthServiceAndGetAuthToken(t, fakeEmail, fakePassword)

	hdlr, err := ConfigureAuthGqlServer(token)
	assert.NoError(t, err)

	c := client.New(hdlr,
		client.AddHeader("Authorization", fmt.Sprintf("Bearer %s", token)))

	query := `
		mutation {
			CreateBusinessAccount(input: {
			    authnId: 10,
			    businessAccount: {
				  companyName: "%s"
			      companyAddress: "340 Clifton Pl"
			      category: "small business",
			      password: "Granada123"
			      email: "%s"
			      isActive: false
			      businessGoals: ["make money", "meet potential clients"]
			      businessStage: "early stage business"
			    }
			  }){
			    id
				isActive
			  }
		}
	`

	resp, err := c.RawPost(fmt.Sprintf(query, fakeCompanyName, fakeEmail))
	ExpectedNoErrorToOccur(t, err, resp)
	ExpectedProperAccountRecord(t, resp, CreateBusinessAccountOperationName)
}

// CreateExistentActiceAccount attempts to create an account that already exists
func CreateExistentActiveAccount(t *testing.T) {
	account := testBusinessAccount
	RandomizeAccount(account)
	ctx := context.TODO()

	token, authnId := graphql_api.CreateAccountInAuthServiceAndGetAuthToken(t, account.Email, account.Password)
	account.AuthnId = authnId

	// save the account
	createdAccount, err := db.CreateBusinessAccount(ctx, account, authnId)
	assert.NoError(t, err)
	assert.NotNil(t, createdAccount)

	hdlr, err := ConfigureAuthGqlServer(token)
	assert.NoError(t, err)

	c := client.New(hdlr,
		client.AddHeader("Authorization", fmt.Sprintf("Bearer %s", token)))

	// generate query attempting to create the same business account
	query := `
		mutation {
			CreateBusinessAccount(input: {
			    authnId: %d,
			    businessAccount: {
				  companyName: "%s"
			      companyAddress: "340 Clifton Pl"
			      category: "small business",
			      password: "%s"
			      email: "%s"
			      isActive: false
			      businessGoals: ["make money", "meet potential clients"]
			      businessStage: "early stage business"
			    }
			  }){
			    id
				isActive
			  }
		}
	`

	q := fmt.Sprintf(query, createdAccount.AuthnId, createdAccount.CompanyName, createdAccount.Password, createdAccount.Email)

	resp, err := c.RawPost(q)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Nil(t, resp.Data)
	assert.NotEmpty(t, resp.Errors)
}

// CreateExistentInactiveAccount activates a previously existent account via distributed transactions
func CreateExistentInactiveAccount(t *testing.T) {
	/*
		In this integration test, we define a test account.
		1. We create an account record from the context of the authentication handler service in the authentication service.
		2. From the previous operation, we obtain the "authn" id as well as a jwt token. We place this token into an "auth context"
		3. To setup the preconditions under which this test needs, we save the account record in the business account service database.
		4. We lock this account (meaning we set it as inactive)
		5. We lock this account from the perspective of the authentication handler service
		6. We wrap the auth context in a custom middleware to be ran prior to the mutation handler
		7. we generate the query string and perform the operation. As a post condition, the account should be reactivated
	*/
	account := testBusinessAccount
	RandomizeAccount(account)
	ctx := context.TODO()

	token, authnId := graphql_api.CreateAccountInAuthServiceAndGetAuthToken(t, account.Email, account.Password)
	account.AuthnId = authnId

	// save the account
	createdAccount, err := db.CreateBusinessAccount(ctx, account, authnId)
	assert.NoError(t, err)
	assert.NotNil(t, createdAccount)

	// set the account as inactive
	err = db.ArchiveBusinessAccount(ctx, createdAccount.Id)
	assert.NoError(t, err)

	// lock the account from the context of the authentication service
	err = graphql_api.LockAccountInAuthService(t, createdAccount.AuthnId, token)
	assert.NoError(t, err)

	hdlr, err := ConfigureAuthGqlServer(token)
	assert.NoError(t, err)

	c := client.New(hdlr,
		client.AddHeader("Authorization", fmt.Sprintf("Bearer %s", token)))

	// generate query attempting to create the same business account
	query := `
		mutation {
			CreateBusinessAccount(input: {
			    authnId: %d,
			    businessAccount: {
				  companyName: "%s"
			      companyAddress: "340 Clifton Pl"
			      category: "small business",
			      password: "%s"
			      email: "%s"
			      isActive: false
			      businessGoals: ["make money", "meet potential clients"]
			      businessStage: "early stage business"
			    }
			  }){
			    id
				isActive
			  }
		}
	`

	q := fmt.Sprintf(query, account.AuthnId, account.CompanyName, account.Password, account.Email)

	resp, err := c.RawPost(q)
	ExpectedNoErrorToOccur(t, err, resp)
	ExpectedProperAccountRecord(t, resp, CreateBusinessAccountOperationName)
}

// Configures a gql handler with custom authentication middleware for tests
func ConfigureAuthGqlServer(token string) (*handler.Server, error) {
	mockAuthMw := func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		ctx = middleware.InjectContextWithMockToken(ctx, token, "test")
		nextCall, err := next(ctx)
		return nextCall, err
	}

	resolvers := graphql_api.Resolver{Db: db}
	hdlr := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers}))
	hdlr.AroundFields(mockAuthMw)
	return hdlr, nil
}

// ExpectedProperAccountRecord ensures the account obtained has the proper fields and no error occurs
func ExpectedProperAccountRecord(t *testing.T, resp *client.Response, operationName string) {
	if resp != nil && resp.Data != nil {
		result := resp.Data.(map[string]interface{})
		// get obtained record
		record := result[operationName].(map[string]interface{})
		assert.NotNil(t, record)
		id := record["id"]
		assert.NotNil(t, id)
		isActive := record["isActive"].(bool)
		assert.True(t, isActive)
	} else {
		t.Fatal("invalid response")
	}
}
