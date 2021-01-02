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
	UpdateBusinessAccountOperationName = "UpdateBusinessAccount"
)

func TestE2EUpdateAccount(t *testing.T) {
	t.Run("TestName:E2E_UpdateExistentAccount", UpdateExistentAccount)
	t.Run("TestName:E2E_UpdateEmailOfExistentAccount", UpdateEmailOfExistentAccount)
	t.Run("TestName:E2E_UpdatePasswordOfExistentAccount", UpdatePasswordOfExistentAccount)
	t.Run("TestName:E2E_UpdateNonExistentAccount", UpdateNonExistentAccount)
}

func UpdateExistentAccount(t *testing.T) {
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

	query := `
		mutation {
			UpdateBusinessAccount(input: {
			    id: %d,
			    businessAccount: {
				  companyName: "%s"
			      email: "%s"
				  companyAddress: "340 Clifton Pl"
			      category: "small business",
			      password: "%s"
			      isActive: true
			      businessGoals: ["make money", "meet potential clients"]
			      businessStage: "early stage business"
			    }
			  }){
			    id
				companyName
				isActive
			  }
		}
	`
	newCompanyName := "test-random-company"
	q := fmt.Sprintf(query, createdAccount.Id, newCompanyName, createdAccount.Email, account.Password)
	resp, err := c.RawPost(q)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.NotNil(t, resp.Data)

	if resp != nil && resp.Data != nil {
		result := resp.Data.(map[string]interface{})
		// get obtained record
		record := result["UpdateBusinessAccount"].(map[string]interface{})
		assert.NotNil(t, record)
		opResult := record["isActive"].(bool)
		assert.True(t, opResult)
		id := record["id"]
		assert.NotNil(t, id)
		companyName := record["companyName"].(string)
		assert.Equal(t, companyName, newCompanyName)
	} else {
		t.Fatal("invalid response")
	}
}

func UpdateEmailOfExistentAccount(t *testing.T) {
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

	query := `
		mutation {
			UpdateBusinessAccount(input: {
			    id: %d,
			    businessAccount: {
				  companyName: "%s"
			      email: "%s"
				  companyAddress: "340 Clifton Pl"
			      category: "small business",
			      password: "%s"
			      isActive: true
			      businessGoals: ["make money", "meet potential clients"]
			      businessStage: "early stage business"
			    }
			  }){
			    id
				email
				isActive
			  }
		}
	`
	newEmail := "test-random-company@gmail.com"
	q := fmt.Sprintf(query, createdAccount.Id, createdAccount.CompanyName, newEmail, account.Password)
	resp, err := c.RawPost(q)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.NotNil(t, resp.Data)

	if resp != nil && resp.Data != nil {
		result := resp.Data.(map[string]interface{})
		// get obtained record
		record := result["UpdateBusinessAccount"].(map[string]interface{})
		assert.NotNil(t, record)
		opResult := record["isActive"].(bool)
		assert.True(t, opResult)
		id := record["id"]
		assert.NotNil(t, id)
		email := record["email"].(string)
		assert.Equal(t, email, newEmail)
	} else {
		t.Fatal("invalid response")
	}
}

func UpdatePasswordOfExistentAccount(t *testing.T) {
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

	query := `
		mutation {
			UpdateBusinessAccount(input: {
			    id: %d,
			    businessAccount: {
				  companyName: "%s"
			      email: "%s"
				  companyAddress: "340 Clifton Pl"
			      category: "small business",
			      password: "%s"
			      isActive: true
			      businessGoals: ["make money", "meet potential clients"]
			      businessStage: "early stage business"
			    }
			  }){
			    id
				email
				isActive
			  }
		}
	`
	password := "test-random-company-passowrd"
	q := fmt.Sprintf(query, createdAccount.Id, createdAccount.CompanyName, createdAccount.Email, password)
	resp, err := c.RawPost(q)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Nil(t, resp.Data)
	assert.NotEmpty(t, resp.Errors)
}

func UpdateNonExistentAccount(t *testing.T) {
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

	query := `
		mutation {
			UpdateBusinessAccount(input: {
			    id: %d,
			    businessAccount: {
				  companyName: "%s"
			      email: "%s"
				  companyAddress: "340 Clifton Pl"
			      category: "small business",
			      password: "%s"
			      isActive: true
			      businessGoals: ["make money", "meet potential clients"]
			      businessStage: "early stage business"
			    }
			  }){
			    id
				email
				isActive
			  }
		}
	`
	password := "test-random-company-passowrd"
	q := fmt.Sprintf(query, 10000, createdAccount.CompanyName, createdAccount.Email, password)
	resp, err := c.RawPost(q)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Nil(t, resp.Data)
	assert.NotEmpty(t, resp.Errors)
}
