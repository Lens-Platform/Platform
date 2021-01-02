package database

import (
	"context"
	"os"
	"time"

	core_database "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-database"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/utils"

	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/saga"
)

type IDbOperations interface {
	// CreateShopperAccount creates a shopper account
	CreateShopperAccount(ctx context.Context, account *model.ShopperAccount, authnId uint32) (*model.ShopperAccount, error)
	// UpdateShopperAccount updates a shopper account by Id
	UpdateShopperAccount(ctx context.Context, id uint32, account *model.ShopperAccount) (*model.ShopperAccount, error)
	// ArchiveShopperAccount archives a shopper account by Id
	ArchiveShopperAccount(ctx context.Context, id uint32) (bool, error)
	// GetShopperAccount gets a shopper account by id
	GetShopperAccount(ctx context.Context, id uint32) (*model.ShopperAccount, error)
	// GetShopperAccounts gets a set of business accounts by id
	GetShopperAccounts(ctx context.Context, limit int) ([]*model.ShopperAccount, error)
	// GetShopperAccountsSubscribedToTopic gets a set of account who've subscribed to a given topic
	GetShopperAccountsSubscribedToTopic(ctx context.Context, topic model.SubscribedTopic) ([]*model.ShopperAccount, error)
}

// Db witholds connection to a postgres database as well as a logging handler
type Db struct {
	// represents database connection
	Conn                                     *core_database.DatabaseConn
	// represents logging utility
	Logger                                   core_logging.ILog
	// tracing utility object
	TracingEngine                            *core_tracing.TracingEngine
	// metrics emitting utility object
	MetricsEngine                            *core_metrics.CoreMetricsEngine
	// authentication handler service connection endpoint
	AuthenticationHandlerServiceBaseEndpoint string
	// saga coordinator
	Saga                                     *saga.SagaCoordinator
	// max allowed connection attempts to the database
	MaxConnectionAttempts                    int
	// max operation retries allowed
	MaxRetriesPerOperation                   int
	// max retry timeout allowed
	RetryTimeOut                             time.Duration
	// max sleep interval between operations that are retried
	OperationSleepInterval                   time.Duration
}

// database operation types
const (
	DbConnectionAttempt = "DB_CONNECTION_ATTEMPT"
)

// DatabaseConfigs database configurations settings
type Configs struct {
	// database connection string
	ConnectionString                         string
	// database object logger
	Logger                                   core_logging.ILog
	// database object tracer
	TracingEngine                            *core_tracing.TracingEngine
	// database object metrics
	MetricsEngine                            *core_metrics.CoreMetricsEngine
	// database authentication handler service endpoint
	AuthenticationHandlerServiceBaseEndpoint string
	// max connection retries configuration
	MaxConnectionAttempts                    int
	// max retry per operation configuration
	MaxRetriesPerOperation                   int
	// max retry timeout configuration
	RetryTimeOut                             time.Duration
	// max sleep interval configuration
	OperationSleepInterval                   time.Duration
}

var (
	modelsToMigrate = []interface{}{
		&model.ShopperAccountORM{},
		&model.AddressORM{},
		&model.CreditCardORM{},
		&model.ImageORM{},
		&model.SubscribedTopicORM{},
		&model.TopicORM{},
	}
)

// New creates a database connection and returns the connection object
func New(ctx context.Context, configs *Configs) (*Db, error) {
	if FaultyConfigs(configs) {
		os.Exit(1)
	}

	logger := configs.Logger
	tracingEngine := configs.TracingEngine
	connectionString := configs.ConnectionString
	metricsEngine := configs.MetricsEngine

	var endpoint = configs.AuthenticationHandlerServiceBaseEndpoint
	if endpoint == "" {
		endpoint = "http://authentication-handler-service:9898/v1/account"
	}

	maxConnectionAttempts := configs.MaxConnectionAttempts
	maxRetriesPerOperation := configs.MaxRetriesPerOperation
	retryTimeout := configs.RetryTimeOut
	operationSleepInterval := configs.OperationSleepInterval

	// generate a span for the database connection
	ctx, span := utils.StartRootOperationSpan(ctx, DbConnectionAttempt, tracingEngine, logger)
	defer span.Finish()

	logger.Info("Attempting database connection operation")
	dbConn := core_database.NewDatabaseConn(connectionString, "postgres")
	if dbConn == nil {
		logger.FatalM(svcErrors.ErrFailedToConnectToDatabase, svcErrors.ErrFailedToConnectToDatabase.Error())
	}
	logger.Info("Successfully connected to the database")

	// configure db
	logger.Info("Attempting database connection configuration")
	ConfigureDatabaseConnection(dbConn)
	logger.Info("Successfully configured database connection object")

	logger.Info("Attempting database schema migration")
	err := MigrateSchemas(dbConn, logger, modelsToMigrate...)
	if err != nil {
		logger.FatalM(err, svcErrors.ErrFailedToPerformDatabaseMigrations.Error())
	}
	logger.Info("Successfully migrated database")

	return &Db{
		Conn:                                     dbConn,
		Logger:                                   logger,
		TracingEngine:                            tracingEngine,
		MetricsEngine:                            metricsEngine,
		AuthenticationHandlerServiceBaseEndpoint: endpoint,
		Saga:                                     saga.NewSagaCoordinator(logger),
		MaxConnectionAttempts:                    maxConnectionAttempts,
		MaxRetriesPerOperation:                   maxRetriesPerOperation,
		RetryTimeOut:                             retryTimeout,
		OperationSleepInterval:                   operationSleepInterval,
	}, nil
}

// FaultyConfigs asserts faulty database configurations were not passed in
func FaultyConfigs(configs *Configs) bool {
	return configs == nil ||
		configs.ConnectionString == utils.EMPTY ||
		configs.TracingEngine == nil ||
		configs.MetricsEngine == nil ||
		configs.Logger == nil ||
		configs.AuthenticationHandlerServiceBaseEndpoint == "" ||
		configs.MaxRetriesPerOperation == 0 ||
		configs.MaxConnectionAttempts == 0 ||
		configs.OperationSleepInterval == 0 ||
		configs.RetryTimeOut == 0
}

// ConfigureDatabaseConnection configures a database connection
func ConfigureDatabaseConnection(dbConn *core_database.DatabaseConn) {
	dbConn.Engine.FullSaveAssociations = true
	dbConn.Engine.SkipDefaultTransaction = false
	dbConn.Engine.PrepareStmt = true
	dbConn.Engine.DisableAutomaticPing = false
	dbConn.Engine = dbConn.Engine.Set("gorm:auto_preload", true)
}

// MigrateSchemas creates or updates a given set of model based on a schema
// if it does not exist or migrates the model schemas to the latest version
func MigrateSchemas(db *core_database.DatabaseConn, logger core_logging.ILog, models ...interface{}) error {
	if err := db.Engine.AutoMigrate(models...); err != nil {
		// TODO: emit metric
		logger.ErrorM(err, svcErrors.ErrFailedToPerformDatabaseMigrations.Error())
		return err
	}

	return nil
}
