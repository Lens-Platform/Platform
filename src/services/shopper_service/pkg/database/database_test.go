package database_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	core_database "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-database"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/database"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
)

var (
	db       *database.Db
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
	testAccount = &model.ShopperAccount{
		Id:                  0,
		FirstName:           "D Yoan L",
		LastName:            "Mekontchou Yomba",
		Email:               "yoanyombapro@gmail.com",
		Username:            "yosieyo",
		Phone:               "424-410-6123",
		IsActive:            false,
		AcceptsMarketing:    false,
		AcceptedMarketingAt: "",
		Addresses:           []*model.Address{
			{
				Id:           0,
				Street:       "340 Clifton Pl",
				Province:     "Brooklyn",
				City:         "NYC",
				ZipCode:      "11013",
				Country:      "USA",
				CountryCode:  "01",
				ProvinceCode: "",
			},
		},
		Tags:                []string{
			"common", "new-user",
		},
		Causes:              []*model.SubscribedTopic{
			{
				Id:              0,
				SubscribedTopic: &model.Topic{
					Id:              0,
					Name:            "Animal Cruelty",
					Tags:            []string{"save animals", "dog", "pets",},
					TopicCoverImage: &model.Image{
						Id:       0,
						Metadata: &model.ImageMeta{
							Width:     300,
							Height:    150,
							RedData:   nil,
							GreenData: nil,
							BlueData:  nil,
						},
						BlobUrl:  "https://amazon.com/akjdsbjksfjdbs/z:2093718372",
						AltText:  "Animal Cruelty",
					},
				},
				SubscribedAt:    "",
				Description:     `
					Cruelty to animals, also called animal abuse, animal neglect or animal cruelty, is the infliction by omission (neglect) or by commission by
					humans of suffering or harm upon any non-human animal. More narrowly, it can be the causing of harm or suffering for specific achievement,
					such as killing animals for entertainment; cruelty to animals sometimes encompasses inflicting harm or suffering as an end in itself, defined as zoosadism.
				`,
			},
		},
		CreditCard:          &model.CreditCard{
			Id:                       0,
			CardNumber:               "1111-2333-3934-3535",
			CardBrand:                "amex",
			ExpiresSoon:              false,
			ExpirationMonth:          5,
			ExpirationYear:           2023,
			FirstDigits:              "1111",
			LastDigits:               "3535",
			MaskedNumber:             "xxxx-xxxx-xxxx-3535",
			CardHolderName:           "Yoan Yomba",
			CreditCardBillingAddress: &model.Address{
				Id:           0,
				Street:       "340 Clifton Pl",
				Province:     "Brooklyn",
				City:         "NYC",
				ZipCode:      "10013",
				Country:      "USA",
				CountryCode:  "01",
				ProvinceCode: "",
			},
		},
		SubscribedTopics:    []*model.SubscribedTopic{
			{
				Id:              0,
				SubscribedTopic: nil,
				SubscribedAt:    "",
				Description:     "empty subscribed topic",
			},
		},
		AuthnId:             0,
		Password:            "Granada123",
	}
)

func TestMain(m *testing.M) {
	const serviceName string = "test"
	const collectorEndpoint = "http://jaeger-collector:14268/api/traces"
	const authenticationHandlerServiceEndpoint = "http://localhost:9898/v1/account"
	// initiate tracing engine
	tracerEngine, closer := core_tracing.NewTracer(serviceName, collectorEndpoint, prometheus.New())
	defer closer.Close()

	// initiate metrics engine
	metricsEngine := core_metrics.NewCoreMetricsEngineInstance("test", nil)

	// initiate logging client
	logger := core_logging.JSONLogger

	// initialize database connection string
	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	ctx := context.Background()
	conn, err := database.New(ctx, &database.Configs{
		ConnectionString:                         connectionString,
		Logger:                                   logger,
		TracingEngine:                            tracerEngine,
		MetricsEngine:                            metricsEngine,
		AuthenticationHandlerServiceBaseEndpoint: authenticationHandlerServiceEndpoint,
		MaxConnectionAttempts:                    5,
		MaxRetriesPerOperation:                   5,
		RetryTimeOut:                             time.Second,
		OperationSleepInterval:                   time.Second,
	})

	db = conn
	if err != nil {
		panic("error - failed to connect to database")
	}

	_ = m.Run()
	return
}

// GenerateRandomizedAccount generates a random account
func GenerateRandomizedAccount() *model.ShopperAccount {
	randStr := core_database.GenerateRandomString(50)
	account := testAccount
	account.Email = account.Email + randStr
	account.Username = account.Username + randStr
	return account
}

// GenerateRandomNumber generates a pseudo random numebr
func GenerateRandomNumber() uint32 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return uint32(r1.Intn(100000))
}

// DefineInitialConditions defines initial conditions prior to running tests
func DefineInitialConditions() (*model.ShopperAccount, uint32, context.Context) {
	testShopperAccount := GenerateRandomizedAccount()
	var authnID uint32 = GenerateRandomNumber()
	ctx := context.TODO()
	return testShopperAccount, authnID, ctx
}
