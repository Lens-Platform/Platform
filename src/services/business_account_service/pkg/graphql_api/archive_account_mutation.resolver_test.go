package graphql_api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/stretchr/testify/assert"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api"
)

var (
	ArchiveBusinessAccountOperationName = "ArchiveBusinessAccount"
)

func TestE2EArchiveAccount(t *testing.T) {
	// archive existent business account
	t.Run("TestName:E2E_ArchiveExistentAccount", ArchiveExistentAccount)
	// archive non existent business account
	t.Run("TestName:E2E_ArchiveNonexistentAccount", ArchiveNonexistentAccount)
}

func ArchiveExistentAccount(t *testing.T) {
	account := testBusinessAccount
	RandomizeAccount(account)
	ctx := context.TODO()

	// create the account in the authentication handler service
	token, authnId := graphql_api.CreateAccountInAuthServiceAndGetAuthToken(t, account.Email, account.Password)

	// create the account locally
	// save the account
	createdAccount, err := db.CreateBusinessAccount(ctx, account, authnId)
	assert.NoError(t, err)
	assert.NotNil(t, createdAccount)

	hdlr, err := ConfigureAuthGqlServer(token)
	assert.NoError(t, err)

	c := client.New(hdlr,
		client.AddHeader("Authorization", fmt.Sprintf("Bearer %s", token)))

	query := `
		mutation {
			DeleteBusinessAccount(id: {
			    id: %d,
			  }){
				result
			}
		}
	`

	resp, err := c.RawPost(fmt.Sprintf(query, createdAccount.Id))
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Errors)

	if resp != nil && resp.Data != nil {
		result := resp.Data.(map[string]interface{})
		// get obtained record
		record := result["DeleteBusinessAccount"].(map[string]interface{})
		assert.NotNil(t, record)
		opResult := record["result"].(bool)
		assert.True(t, opResult)
	} else {
		t.Fatal("invalid response")
	}
}

func ArchiveNonexistentAccount(t *testing.T) {
	account := testBusinessAccount
	RandomizeAccount(account)
	ctx := context.TODO()

	// create the account in the authentication handler service
	token, authnId := graphql_api.CreateAccountInAuthServiceAndGetAuthToken(t, account.Email, account.Password)

	// create the account locally
	// save the account
	createdAccount, err := db.CreateBusinessAccount(ctx, account, authnId)
	assert.NoError(t, err)
	assert.NotNil(t, createdAccount)

	hdlr, err := ConfigureAuthGqlServer(token)
	assert.NoError(t, err)

	c := client.New(hdlr,
		client.AddHeader("Authorization", fmt.Sprintf("Bearer %s", token)))

	query := `
		mutation {
			DeleteBusinessAccount(id: {
			    id: %d,
			  }){
				result
			}
		}
	`

	// we pass an account id that can't exist
	resp, err := c.RawPost(fmt.Sprintf(query, 10000))
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Nil(t, resp.Data)
	assert.NotEmpty(t, resp.Errors)
}
