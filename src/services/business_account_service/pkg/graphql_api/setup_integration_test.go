package graphql_api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/database"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/model"
)

var (
	db                  *database.Db
	host                = "localhost"
	port                = 5433
	user                = "postgres"
	password            = "postgres"
	dbname              = "postgres"
	testBusinessAccount = &model.BusinessAccount{
		Id: 0,
		CompanyName:                 "BlackspaceInc",
		CompanyAddress:              "340 Clifton Pl",
		PhoneNumber:                 &model.PhoneNumber{
			Number: "424-410-6123",
			Type:   &model.PhoneType{
				Home:   true,
				Work:   false,
				Mobile: false,
			},
		},
		Category:                    "small business",
		Media:                       &model.Media{
			Id:        0,
			Website:   "space.com",
			Instagram: "space@instagram.com",
			Facebook:  "space@facebook.com",
			LinkedIn:  "space@linkedin.com",
			Pinterest: "space@pinterest.com",
		},
		Password:                    "Granada123",
		Email:                       "space@gmail.com",
		IsActive:                    false,
		TypeOfBusiness:              &model.BusinessType{
			Category:    &model.BusinessCategory{
				Tech:                         true,
				CharitiesEducationMembership: false,
				FoodAndDrink:                 false,
				HealthCareAndFitness:         true,
				HomeAndRepair:                false,
				LeisureAndEntertainment:      false,
				ProfessionalServices:         false,
				Retail:                       false,
				Transportation:               false,
				BeautyAndPersonalCare:        false,
			},
			SubCategory: &model.BusinessSubCategory{
				Marketing:         false,
				Travel:            false,
				Interior_Design:   false,
				Music:             false,
				Technology:        true,
				Food:              false,
				Restaurants:       false,
				Polictics:         false,
				Health_And_Beauty: false,
				Design:            false,
				Non_Profit:        false,
				Jewelry:           false,
				Gaming:            false,
				Magazine:          false,
				Photography:       false,
				Fitenss:           true,
				Consulting:        false,
				Fashion:           false,
				Services:          false,
				Art:               false,
			},
		},
		BusinessGoals:               []string{"make a lot of money", "change the world"},
		BusinessStage:               "small business",
		MerchantType:                &model.MerchantType{
			SoleProprietor:        true,
			SideProject:           false,
			CasualUse:             false,
			LLCCorporation:        true,
			Partnership:           false,
			Charity:               false,
			ReligiousOrganization: false,
			OnePersonBusiness:     true,
		},
		PaymentDetails:              &model.PaymentProcessingMethods{
			PaymentOptions: []*model.PaymentOptions{{
				BrickAndMortar:  true,
				OnTheGo:         true,
				Online:          false,
				ThroughInvoices: false,
			}},
			Medium:         nil,
		},
		ServicesManagedByBlackspace: nil,
		FounderAddress:              nil,
		SubscribedTopics:            nil,
		AuthnId:                     0,
	}
)

func InitializeDatabaseConnection() *database.Db {
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

	return database.Setup(ctx, connectionString, tracerEngine, serviceMetrics, logger, "http://localhost:9898/v1/account")
}

func TestMain(m *testing.M) {
	db = InitializeDatabaseConnection()

	// defer deleting all created entries
	// cleanupHandler := database.DeleteCreatedEntities(db.Engine)
	// defer cleanupHandler()

	_ = m.Run()
	return
}
