package database

import (
	"context"
	"os"
	"time"

	core_database "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-database"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"gorm.io/gorm"

	svcErrors "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/model"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/saga"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/utils"
)

// IDatabase provides an interface which any database tied to this service should implement
type IDatabase interface {
	// To be implemented
}

// Db witholds connection to a postgres database as well as a logging handler
type Db struct {
	Conn                                     *core_database.DatabaseConn
	Logger                                   core_logging.ILog
	TracingEngine                            *core_tracing.TracingEngine
	MetricsEngine                            *core_metrics.CoreMetricsEngine
	AuthenticationHandlerServiceBaseEndpoint string
	Saga                                     *saga.SagaCoordinator
	MaxConnectionAttempts                    int
	MaxRetriesPerOperation                   int
	RetryTimeOut                             time.Duration
	OperationSleepInterval                   time.Duration
}

// Tx is a type serving as a function decorator for common database transactions
type Tx func(ctx context.Context, tx *gorm.DB) error

// CmplxTx is a type serving as a function decorator for complex database transactions
type CmplxTx func(ctx context.Context, tx *gorm.DB) (interface{}, error)

// database operation types
const (
	DB_CONNECTION_ATTEMPT = "DB_CONNECTION_ATTEMPT"
)

// New creates a database connection and returns the connection object
func New(ctx context.Context, connectionString string, tracingEngine *core_tracing.TracingEngine, metricsEngine *core_metrics.CoreMetricsEngine,
	logger core_logging.ILog, svcEndpoint string) (*Db,
	error) {
	// generate a span for the database connection
	ctx, span := utils.StartRootOperationSpan(ctx, DB_CONNECTION_ATTEMPT, tracingEngine, logger)
	defer span.Finish()

	if connectionString == utils.EMPTY || tracingEngine == nil || metricsEngine == nil || logger == nil {
		// crash the process
		os.Exit(1)
	}

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
	err := MigrateSchemas(dbConn, logger, &model.BusinessAccountORM{}, &model.MediaORM{}, &model.TopicsORM{})
	if err != nil {
		logger.FatalM(err, svcErrors.ErrFailedToPerformDatabaseMigrations.Error())
	}
	logger.Info("Successfully migrated database")

	var endpoint = svcEndpoint
	if endpoint == "" {
		endpoint = "http://authentication-handler-service:9898/v1/account"
	}

	return &Db{
		Conn:                                     dbConn,
		Logger:                                   logger,
		TracingEngine:                            tracingEngine,
		MetricsEngine:                            metricsEngine,
		AuthenticationHandlerServiceBaseEndpoint: svcEndpoint,
		Saga:                                     saga.NewSagaCoordinator(logger),
		MaxConnectionAttempts:                    5,
		MaxRetriesPerOperation:                   3,
		RetryTimeOut:                             1 * time.Millisecond,
		OperationSleepInterval:                   1 * time.Second,
	}, nil
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
