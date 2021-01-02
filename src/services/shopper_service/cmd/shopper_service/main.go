package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	core_auth_sdk "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-auth-sdk"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/database"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/grpc"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/signals"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/version"
)

func main() {
	// flags definition
	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)
	fs.Int("port", 9896, "HTTP port")
	fs.Int("secure-port", 0, "HTTPS port")
	fs.Int("port-metrics", 9796, "metrics port")
	fs.Int("grpc-port", 0, "gRPC port")
	fs.String("grpc-service-name", "shopper_service", "gPRC service name")
	fs.String("level", "info", "log level debug, info, warn, error, flat or panic")
	fs.StringSlice("backend-url", []string{}, "backend service URL")
	fs.Duration("http-client-timeout", 2*time.Minute, "client timeout duration")
	fs.Duration("http-server-timeout", 30*time.Second, "server read and write timeout duration")
	fs.Duration("http-server-shutdown-timeout", 5*time.Second, "server graceful shutdown timeout duration")
	fs.String("data-path", "/data", "data local path")
	fs.String("config-path", "", "config dir path")
	fs.String("cert-path", "/data/cert", "certificate path for HTTPS port")
	fs.String("config", "config.yaml", "config file name")
	fs.String("ui-path", "./ui", "UI local path")
	fs.String("ui-logo", "", "UI logo")
	fs.String("ui-color", "#34577c", "UI color")
	fs.String("ui-message", fmt.Sprintf("greetings from shopper_service v%v", version.VERSION), "UI message")
	fs.Bool("h2c", false, "allow upgrading to H2C")
	fs.Bool("random-delay", false, "between 0 and 5 seconds random delay by default")
	fs.String("random-delay-unit", "s", "either s(seconds) or ms(milliseconds")
	fs.Int("random-delay-min", 0, "min for random delay: 0 by default")
	fs.Int("random-delay-max", 5, "max for random delay: 5 by default")
	fs.Bool("random-error", false, "1/3 chances of a random response error")
	fs.Bool("unhealthy", false, "when set, healthy state is never reached")
	fs.Bool("unready", false, "when set, ready state is never reached")
	fs.Int("stress-cpu", 0, "number of CPU cores with 100 load")
	fs.Int("stress-memory", 0, "MB of data to load into memory")
	fs.String("cache-server", "", "Redis address in the format <host>:<port>")

	// service configurations
	fs.String("service_name", "shopper_account_service", "microservice name")

	// authentication service configurations
	fs.String("authentication_handler_service_base", "http://authentication_handler_service",
		"authentication handler service endpoint base address")
	fs.Int("authentication_handler_service_port", 9898,
		"authentication handler service endpoint port")
	fs.String("authentication_handler_service_sub_address", "/v1/account",
		"authentication handler service endpoint base address")

	// tracing configurations
	fs.String("jaeger-endpoint", "http://jaeger-collector:14268/api/traces", "jaeger collector endpoint")

	// database connection configurations
	fs.String("db_host", "shopper_account_service_db", "database host string")
	fs.Int("db_port", 5432, "database port")
	fs.String("db_user", "postgres", "database user string")
	fs.String("db_password", "postgres", "database password string")
	fs.String("db_name", "postgres", "database name")

	// authn client connection
	fs.String("authn_username", "blackspaceinc", "username of authentication client")
	fs.String("authn_password", "blackspaceinc", "password of authentication client")
	fs.String("authn_issuer_base_url", "http://localhost", "authentication service issuer")
	fs.String("authn_origin", "http://localhost", "origin of auth requests")
	fs.String("authn_domains", "localhost", "authentication service domains")
	fs.String("authn_private_base_url", "http://authentication_service",
		"authentication service private url. should be local host if these are not running on docker containers. "+
			"However if running in docker container with a configured docker network, the url should be equal to the service name")
	fs.String("authn_public_base_url", "http://localhost", "authentication service public endpoint")
	fs.String("authn_internal_port", "3000", "authentication service port")
	fs.String("authn_port", "8404", "authentication service external port")

	versionFlag := fs.BoolP("version", "v", false, "get version number")

	// parse flags
	err := fs.Parse(os.Args[1:])
	switch {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		fs.PrintDefaults()
		os.Exit(2)
	case *versionFlag:
		fmt.Println(version.VERSION)
		os.Exit(0)
	}

	// bind flags and environment variables
	viper.BindPFlags(fs)
	viper.RegisterAlias("backendUrl", "backend-url")
	hostname, _ := os.Hostname()
	viper.SetDefault("jwt-secret", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	viper.SetDefault("ui-logo", "https://raw.githubusercontent.com/stefanprodan/shopper_service/gh-pages/cuddle_clap.gif")
	viper.Set("hostname", hostname)
	viper.Set("version", version.VERSION)
	viper.Set("revision", version.REVISION)
	viper.SetEnvPrefix("shopper_account_service")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// load config from file
	if _, fileErr := os.Stat(filepath.Join(viper.GetString("config-path"), viper.GetString("config"))); fileErr == nil {
		viper.SetConfigName(strings.Split(viper.GetString("config"), ".")[0])
		viper.AddConfigPath(viper.GetString("config-path"))
		if readErr := viper.ReadInConfig(); readErr != nil {
			fmt.Printf("Error reading config file, %v\n", readErr)
		}
	} else {
		fmt.Printf("Error to open config file, %v\n", fileErr)
	}

	// configure tracing
	serviceName, tracerEngine, closer := SetupDistributedTracing()
	defer closer.Close()

	if tracerEngine == nil {
		panic("cannot initialize tracer engine")
	}
	opentracing.SetGlobalTracer(tracerEngine.Tracer)

	// configure logging
	ctx := context.Background()
	logger := SetupLogger(ctx)

	// configure metrics
	coreMetrics := core_metrics.NewCoreMetricsEngineInstance(serviceName, nil)

	// sets up database connection
	db, err := SetupDatabaseConnection(logger, err, ctx, tracerEngine, coreMetrics)
	if err != nil {
		logger.For(ctx).FatalM(err, "database connectivity failure")
	}

	// configure authn client connection for middleware auth
	authClient, err := SetupAuthnClient()
	if err != nil {
		logger.For(ctx).FatalM(err, "auth client configuration failure")
	}

	// start stress tests if any
	beginStressTest(viper.GetInt("stress-cpu"), viper.GetInt("stress-memory"), logger)

	// validate port
	if _, err := strconv.Atoi(viper.GetString("port")); err != nil {
		port, _ := fs.GetInt("port")
		viper.Set("port", strconv.Itoa(port))
	}

	// validate secure port
	if _, err := strconv.Atoi(viper.GetString("secure-port")); err != nil {
		securePort, _ := fs.GetInt("secure-port")
		viper.Set("secure-port", strconv.Itoa(securePort))
	}

	// validate random delay options
	if viper.GetInt("random-delay-max") < viper.GetInt("random-delay-min") {
		logger.FatalM(errors.New("invalid random delay configurations"), "`--random-delay-max` should be greater than `--random-delay-min`")
	}

	switch delayUnit := viper.GetString("random-delay-unit"); delayUnit {
	case
		"s",
		"ms":
		break
	default:
		logger.FatalM(errors.New("Invalid random delay configurations"),"`random-delay-unit` accepted values are: s|ms")
	}

	// load gRPC server config
	var grpcCfg grpc.Config
	if err := viper.Unmarshal(&grpcCfg); err != nil {
		logger.FatalM(err,"config unmarshal failed")
	}

	// start gRPC server
	if grpcCfg.Port > 0 {
		grpcSrv, _ := grpc.NewServer(&grpcCfg, logger)
		go grpcSrv.ListenAndServe()
	}

	// load HTTP server config
	var srvCfg api.Config
	if err := viper.Unmarshal(&srvCfg); err != nil {
		logger.FatalM(err, "config unmarshal failed")
	}

	// log version and port
	logger.Info("Starting shopper_service",
		zap.String("version", viper.GetString("version")),
		zap.String("revision", viper.GetString("revision")),
		zap.String("port", srvCfg.Port),
	)

	// start HTTP server
	srv, _ := api.NewServer(&srvCfg, logger, tracerEngine, coreMetrics, db, authClient)
	stopCh := signals.SetupSignalHandler()
	srv.ListenAndServe(stopCh)
}

// SetupDatabaseConnection sets up a database connection
func SetupDatabaseConnection(logger core_logging.ILog, err error, ctx context.Context, tracerEngine *core_tracing.TracingEngine,
	coreMetrics *core_metrics.CoreMetricsEngine)(*database.Db, error) {
	authSvcBase := viper.GetString("authentication_handler_service_base")
	authSvcPort := viper.GetInt("authentication_handler_service_port")
	authSvcSubAddress := viper.GetString("authentication_handler_service_sub_address")
	authSvcEndpoint := fmt.Sprintf("%s:%d%s", authSvcBase, authSvcPort, authSvcSubAddress)

	// configure db connection
	connectionString := configureConnectionString()
	logger.Info(fmt.Sprintf("Database connection string : %s ", connectionString))

	return database.New(ctx, &database.Configs{
		ConnectionString:                         connectionString,
		Logger:                                   logger,
		TracingEngine:                            tracerEngine,
		MetricsEngine:                            coreMetrics,
		AuthenticationHandlerServiceBaseEndpoint: authSvcEndpoint,
		MaxConnectionAttempts:                    5,
		MaxRetriesPerOperation:                   5,
		RetryTimeOut:                             200 * time.Millisecond,
		OperationSleepInterval:                   200 * time.Millisecond,
	})
}

// SetupLogger sets up logging facility
func SetupLogger(ctx context.Context) core_logging.ILog {
	rootSpan := opentracing.SpanFromContext(ctx)
	logger := core_logging.NewJSONLogger(nil, rootSpan)
	return logger
}

// SetupDistributedTracing sets up distributed through jaeger
func SetupDistributedTracing() (string, *core_tracing.TracingEngine, io.Closer) {
	serviceName := viper.GetString("service_name")
	collectorEndpoint := viper.GetString("jaeger-endpoint")
	// initialize a tracing object globally
	tracerEngine, closer := core_tracing.NewTracer(serviceName, collectorEndpoint, prometheus.New())
	return serviceName, tracerEngine, closer
}

var stressMemoryPayload []byte

func beginStressTest(cpus int, mem int, logger core_logging.ILog) {
	done := make(chan int)
	if cpus > 0 {
		logger.Info("starting CPU stress", zap.Int("cores", cpus))
		for i := 0; i < cpus; i++ {
			go func() {
				for {
					select {
					case <-done:
						return
					default:

					}
				}
			}()
		}
	}

	if mem > 0 {
		path := "/tmp/shopper_service.data"
		f, err := os.Create(path)

		if err != nil {
			logger.Error(err, "memory stress failed")
		}

		if err := f.Truncate(1000000 * int64(mem)); err != nil {
			logger.Error(err, "memory stress failed")
		}

		stressMemoryPayload, err = ioutil.ReadFile(path)
		f.Close()
		os.Remove(path)
		if err != nil {
			logger.Error(err,"memory stress failed")
		}
		logger.Info("starting CPU stress", zap.Int("memory", len(stressMemoryPayload)))
	}
}

// configureConnectionString configures database connection string
func configureConnectionString() string {
	host := viper.GetString("db_host")
	port := viper.GetInt("db_port")
	user := viper.GetString("db_user")
	password := viper.GetString("db_password")
	dbname := viper.GetString("db_name")
	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	return connectionString
}

// SetupAuthnClient configures the authentication service client
func SetupAuthnClient() (*core_auth_sdk.Client, error) {
	username := viper.GetString("authn_username")
	password := viper.GetString("authn_password")
	audience := viper.GetString("authn_domains")
	url := viper.GetString("authn_private_base_url") + ":" + viper.GetString("authn_internal_port")
	origin := viper.GetString("authn_origin")
	issuer := viper.GetString("authn_issuer_base_url") + ":" + viper.GetString("authn_port")

	return core_auth_sdk.NewClient(core_auth_sdk.Config{
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
		PrivateBaseURL: url,
	}, origin)
}
