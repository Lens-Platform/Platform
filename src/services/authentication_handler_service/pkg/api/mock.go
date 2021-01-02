package api

import (
	"context"
	"io"
	"time"

	core_auth_sdk "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-auth-sdk"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/metrics"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/middleware"
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

	// authn client
	authnClient, err := InitializeAuthnClient(logger)
	if err != nil {
		logger.For(ctx).FatalM(err, "unable to setup mock server")
	}

	srv := &Server{
		router:        mux.NewRouter(),
		config:        config,
		tracerEngine:  tracerEngine,
		metricsEngine: serviceMetrics.Engine,
		metrics:       serviceMetrics.MicroServiceMetrics,
		logger:        logger,
		authnClient:   authnClient,
	}

	authMw := middleware.NewAuthnMw(srv.authnClient, srv.logger)
	srv.router.Use(authMw.AuthenticationMiddleware)

	return srv
}

func InitializeAuthnClient(logger core_logging.ILog) (*core_auth_sdk.Client, error) {
	// TODO Move this to errors folder
	const username string = "blackspaceinc"
	const password string = "blackspaceinc"
	const audience string = "localhost"
	const issuer string = "http://localhost:8404"
	const origin string = "http://localhost"
	const privateBaseUrl string = "http://localhost:8404"

	client, err := core_auth_sdk.NewClient(core_auth_sdk.Config{
		// The AUTHN_URL of your Keratin AuthN server. This will be used to verify tokens created by
		// AuthN, and will also be used for API calls unless PrivateBaseURL is also set.
		Issuer: issuer,

		// The domain of your application (no protocol). This domain should be listed in the APP_DOMAINS
		// of your Keratin AuthN server.
		Audience: audience,

		// Credentials for AuthN's private endpoints. These will be used to execute admin actions using
		// the Client provided by this library.
		//
		// TIP: make them extra secure in production!
		Username: username,
		Password: password,

		// RECOMMENDED: Send private API calls to AuthN using private network routing. This can be
		// necessary if your environment has a firewall to limit public endpoints.
		PrivateBaseURL: privateBaseUrl,
	}, origin)

	// TODO: make this a retryable operation
	retries := 1
	for retries < 4 {
		// perform a test request to the authentication service
		data, err := client.ServerStats()
		if err != nil {
			if retries != 4 {
				logger.ErrorM(err, "failed to connect to authentication service")
			} else {
				logger.FatalM(err, "failed to connect to authentication service")
			}
			retries += 1
		} else {
			retries = 4
			logger.InfoM("data", zap.Any("result", data))
		}

		time.Sleep(1 * time.Second)
	}

	return client, err
}

func InitializeLoggingEngine(ctx context.Context) core_logging.ILog {
	// initiate authn client
	rootSpan := opentracing.SpanFromContext(ctx)

	// create logging object
	logger := core_logging.NewJSONLogger(nil, rootSpan)
	return logger
}

func InitializeMetricsEngine(serviceName string) *metrics.MetricsEngine {
	coreMetrics := core_metrics.NewCoreMetricsEngineInstance(serviceName, nil)
	serviceMetrics := metrics.NewMetricsEngine(coreMetrics)
	return serviceMetrics
}

func InitializeTracingEngine(serviceName string) (*core_tracing.TracingEngine, io.Closer) {
	// TODO move this to constant folder
	const collectorEndpoint string = "http://localhost:14268/api/traces"

	// initiaize a tracing object globally
	return core_tracing.NewTracer(serviceName, collectorEndpoint, prometheus.New())
}
