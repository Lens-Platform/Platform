package graphql_api_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/assert"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/generated"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/model"
)

func TestE2EGetUserAccount(t *testing.T) {
	t.Run("TestName:E2E_GetNonExistentAccount", GetNonExistentAccount)
	t.Run("TestName:E2E_GetExistingAccount", GetExistingAccount)
	t.Run("TestName:E2E_GetAccountMisconfiguredInput", GetAccountMisconfiguredInput)
}

func TestE2EGetUserAccounts(t *testing.T) {
	t.Run("TestName:E2E_GetExistingUserAccounts", GetExistingUserAccounts)
	t.Run("TestName:E2E_GetNonExistentAccounts", GetNonExistentUserAccounts)
}

func GetExistingUserAccounts(t *testing.T) {
	resolvers := graphql_api.Resolver{Db: db}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})))

	// create a number of accounts
	var numAccountsToCreate int = 10
	for i := 0; i < numAccountsToCreate; i++ {
		ctx := context.TODO()
		account := testBusinessAccount
		RandomizeAccount(account)

		var authnId uint32 = uint32(i + 1)
		account, err := db.CreateBusinessAccount(ctx, account, authnId)
		assert.Empty(t, err)
		assert.NotNil(t, account)
	}

	q := `
		query {
		  getBusinessAccounts(limit: {
		    limit: 5
		  }){
		    companyName,
			email
		  }
		}
	`

	resp, err := c.RawPost(q)
	ExpectedNoErrorToOccur(t, err, resp)
}

func GetNonExistentUserAccounts(t *testing.T) {
	resolvers := graphql_api.Resolver{Db: db}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})))
	q := `
		query {
		  getBusinessAccounts(limit: {
		    limit: 10
		  }){
		    companyName,
			email
		  }
		}
	`

	// since no accounts were created there should be no values returned
	resp, err := c.RawPost(q)
	ExpectedNoErrorToOccur(t, err, resp)
}

func GetNonExistentAccount(t *testing.T) {
	resolvers := graphql_api.Resolver{Db: db}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})))

	q := `
		query {
		  getBusinessAccount(input: {
		    id: 1000
		  }){
		    companyName
		  }
		}
	`
	resp, err := c.RawPost(q)
	ExpectedErrorToOccur(t, err, resp)
}

func GetExistingAccount(t *testing.T) {
	resolvers := graphql_api.Resolver{Db: db}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})))

	var authnId uint32 = 20
	ctx := context.TODO()

	account := testBusinessAccount
	RandomizeAccount(account)

	account, err := db.CreateBusinessAccount(ctx, account, authnId)
	assert.Empty(t, err)
	assert.NotNil(t, account)

	id, err := strconv.Atoi(fmt.Sprint(account.Id))
	assert.Empty(t, err)

	query := fmt.Sprintf(
		`
		query {
		  getBusinessAccount(input: {
		    id: %d
		  }){
			id
			companyName
			password
			email
			isActive
			businessGoals
			businessStage
			authnId
		  }
		}
	`, id)

	resp, err := c.RawPost(query)
	ExpectedNoErrorToOccur(t, err, resp)
}

func ExpectedNoErrorToOccur(t *testing.T, err error, resp *client.Response) {
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Data)
	assert.NoError(t, err)
	assert.Empty(t, resp.Errors)
}

func GetAccountMisconfiguredInput(t *testing.T) {
	resolvers := graphql_api.Resolver{Db: db}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})))
	// test case we have a misconfigured/non-existent user id
	q := `
		query {
		  getBusinessAccount(input: {
		    id: -1
		  }){
		    companyName
		  }
		}
	`

	resp, err := c.RawPost(q)
	ExpectedErrorToOccur(t, err, resp)
}

func ExpectedErrorToOccur(t *testing.T, err error, resp *client.Response) {
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Errors)
	assert.Nil(t, resp.Data)
}

func RandomizeAccount(account *model.BusinessAccount) {
	var randStr = graphql_api.GenerateRandomString(50)
	account.Email = account.Email + randStr
	account.CompanyName = account.CompanyName + randStr
}
