package api

import (
	"context"
	"fmt"
	"io"
	"time"

	gql "github.com/99designs/gqlgen/graphql/handler"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/database"
	graphql "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/generated"
)

var (
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

func NewMockServer() *Server {
	config := &Config{
		Port:                      "9898",
		HttpServerShutdownTimeout: 5 * time.Second,
		HttpServerTimeout:         30 * time.Second,
		BackendURL:                []string{},
		ConfigPath:                "/config",
		DataPath:                  "/data",
		HttpClientTimeout:         30 * time.Second,
		UIColor:                   "blue",
		UIPath:                    ".ui",
		UIMessage:                 "Greetings",
		Hostname:                  "localhost",
	}

	const serviceName string = "test"

	// initiate tracing engine
	tracerEngine, closer := InitializeTracingEngine(serviceName)
	defer closer.Close()
	ctx := context.Background()

	// initiate metrics engine
	serviceMetrics := InitializeMetricsEngine(serviceName)

	// initiate logging client
	logger := InitializeLoggingEngine(ctx)

	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db := database.Setup(ctx, connectionString, tracerEngine, serviceMetrics, logger, "")

	gqlServer := gql.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graphql.Resolver{
		Db:      db,
		Logger:  logger,
		Tracer:  tracerEngine,
		Metrics: serviceMetrics,
	}}))

	srv := &Server{
		router:        mux.NewRouter(),
		config:        config,
		tracingEngine: tracerEngine,
		metricsEngine: serviceMetrics,
		logger:        logger,
		gqlServer:     gqlServer,
	}

	return srv
}

func InitializeLoggingEngine(ctx context.Context) core_logging.ILog {
	// initiate authn client
	rootSpan := opentracing.SpanFromContext(ctx)

	// create logging object
	logger := core_logging.NewJSONLogger(nil, rootSpan)
	return logger
}

func InitializeMetricsEngine(serviceName string) *core_metrics.CoreMetricsEngine {
	return core_metrics.NewCoreMetricsEngineInstance(serviceName, nil)
}

func InitializeTracingEngine(serviceName string) (*core_tracing.TracingEngine, io.Closer) {
	// TODO move this to constant folder
	const collectorEndpoint string = "http://localhost:14268/api/traces"

	// initiaize a tracing object globally
	return core_tracing.NewTracer(serviceName, collectorEndpoint, prometheus.New())
}
