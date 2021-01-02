package database_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/database"
	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/proto"
)

var (
	db       *database.Db
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

var (
	testBusinessAccount = &proto.BusinessAccount{
		Id:             0,
		CompanyName:    "BlackspaceInc",
		CompanyAddress: "340 clifton pl",
		PhoneNumber: &proto.PhoneNumber{
			Number: "424-410-6123",
			Type:   0,
		},
		Category: "small business",
		Media: &proto.Media{
			Id:        0,
			Website:   "blackspaceInc.com",
			Instagram: "blkspace",
			Facebook:  "blkspace_",
			LinkedIn:  "",
			Pinterest: "",
		},
		Password: "Granada123",
		Email:    "BlackspaceInc@gmail.com",
		IsActive: false,
		TypeOfBusiness: &proto.BusinessType{
			Category:    proto.BusinessCategory_Tech,
			SubCategory: proto.BusinessSubCategory_Technology,
		},
		BusinessGoals: []string{"onboard as many customers as possible"},
		BusinessStage: "small business",
		MerchantType:  proto.MerchantType_CasualUse,
		PaymentDetails: &proto.PaymentProcessingMethods{
			PaymentOptions: []proto.PaymentOptions{proto.PaymentOptions_Online},
			Medium:         nil,
		},
		ServicesManagedByBlackspace: proto.ServicesManagedByBlackspace_FundingYourBusiness,
		FounderAddress: &proto.Address{
			Address:       "340 Clifton Pl",
			ApartmentUnit: "3D",
			ZipCode:       "19101",
			City:          "Brooklyn",
			State:         "NY",
			Birthdate: &proto.DateOfBirth{
				Month: "july",
				Day:   "12",
				Year:  "1996",
			},
		},
		SubscribedTopics: &proto.Topics{Business: true},
		AuthnId:          0,
	}
)

func TestMain(m *testing.M) {
	const serviceName string = "test"
	// initiate tracing engine
	tracerEngine, closer := api.InitializeTracingEngine(serviceName)
	defer closer.Close()
	ctx := context.Background()

	// initiate metrics engine
	serviceMetrics := api.InitializeMetricsEngine(serviceName)

	// initiate logging client
	logger := api.InitializeLoggingEngine(ctx)

	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db = database.Setup(ctx, connectionString, tracerEngine, serviceMetrics, logger, "")

	_ = m.Run()
	return
}

// GenerateRandomId generates a random id over a range
func GenerateRandomId(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

//  ExpectNoErrorOccured ensures no errors occured during the operation
func ExpectNoErrorOccured(t *testing.T, err error, result *proto.BusinessAccount) {
	assert.Empty(t, err)
	assert.NotNil(t, result)
}

// ExpectValidAccountObtained ensures we have a valid obtained account
func ExpectValidAccountObtained(t *testing.T, err error, obtainedAccount *proto.BusinessAccount, result *proto.BusinessAccount) {
	assert.Empty(t, err)
	assert.True(t, obtainedAccount != nil)
	assert.Equal(t, obtainedAccount.CompanyName, result.CompanyName)
	assert.Equal(t, obtainedAccount.Email, result.Email)
	assert.Equal(t, obtainedAccount.Password, result.Password)
}

// ExpectInvalidArgumentsError ensure the invalid error is present
func ExpectInvalidArgumentsError(t *testing.T, err error, account *proto.BusinessAccount) {
	assert.NotEmpty(t, err)
	assert.EqualError(t, err, svcErrors.ErrInvalidInputArguments.Error())
	assert.Nil(t, account)
}

// ExpectAccountAlreadyExistError ensures the account already exist error is present
func ExpectAccountAlreadyExistError(t *testing.T, err error, createdAccount *proto.BusinessAccount) {
	assert.NotEmpty(t, err)
	assert.EqualError(t, err, svcErrors.ErrAccountAlreadyExist.Error())
	assert.Nil(t, createdAccount)
}

// ExpectAccountDoesNotExistError ensures the account does not exist error is present
func ExpectAccountDoesNotExistError(t *testing.T, err error, createdAccount *proto.BusinessAccount) {
	assert.NotEmpty(t, err)
	assert.EqualError(t, err, svcErrors.ErrAccountDoesNotExist.Error())
	assert.Nil(t, createdAccount)
}

// ExpectCannotUpdatePasswordError ensure the invalid error is present
func ExpectCannotUpdatePasswordError(t *testing.T, err error, account *proto.BusinessAccount) {
	assert.NotEmpty(t, err)
	assert.EqualError(t, err, svcErrors.ErrCannotUpdatePassword.Error())
	assert.Nil(t, account)
}

// GenerateRandomizedAccount generates a random account
func GenerateRandomizedAccount() *proto.BusinessAccount {
	randStr := database.GenerateRandomString(150)
	account := testBusinessAccount
	account.Email = account.Email + randStr
	account.CompanyName = account.CompanyName + randStr
	return account
}
