package main

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/uber/jaeger-lib/metrics/prometheus"

	// core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	// "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/grpc"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/metrics"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/signals"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version"
)

func main() {
	// flags definition
	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)
	fs.Int("port", 9898, "HTTP port")
	fs.Int("secure-port", 0, "HTTPS port")
	fs.Int("port-metrics", 0, "metrics port")
	fs.Int("grpc-port", 0, "gRPC port")
	fs.String("grpc-service-name", "authentication_handler_service", "gPRC service name")
	fs.String("level", "info", "log level debug, info, warn, error, flat or panic")
	fs.StringSlice("backend-url", []string{}, "backend service URL")
	fs.Duration("HTTP_CLIENT_TIMEOUT", 2*time.Minute, "client timeout duration")
	fs.Duration("http-server-timeout", 30*time.Second, "server read and write timeout duration")
	fs.Duration("http-server-shutdown-timeout", 5*time.Second, "server graceful shutdown timeout duration")
	fs.String("data-path", "/data", "data local path")
	fs.String("config-path", "", "config dir path")
	fs.String("cert-path", "/data/cert", "certificate path for HTTPS port")
	fs.String("config", "config.yaml", "config file name")
	fs.String("ui-path", "./ui", "UI local path")
	fs.String("ui-logo", "", "UI logo")
	fs.String("ui-color", "#34577c", "UI color")
	fs.String("ui-message", fmt.Sprintf("greetings from authentication_handler_service v%v", version.VERSION), "UI message")
	fs.Bool("h2c", false, "allow upgrading to H2C")
	fs.Bool("random-delay", false, "between 0 and 5 seconds random delay by default")
	fs.String("random-delay-unit", "s", "either s(seconds) or ms(milliseconds")
	fs.Int("random-delay-min", 0, "min for random delay: 0 by default")
	fs.Int("random-delay-max", 5, "max for random delay: 5 by default")
	fs.Bool("random-error", false, "1/3 chances of a random response error")
	fs.Bool("unhealthy", false, "when set, healthy state is never reached")
	fs.Bool("unready", false, "when set, ready state is never reached")
	fs.Int("STRESS_CPU", 0, "number of CPU cores with 100 load")
	fs.Int("STRESS_MEMORY", 0, "MB of data to load into memory")
	// authentication service specific flags
	fs.String("AUTHN_USERNAME", "blackspaceinc", "username of authentication client")
	fs.String("AUTHN_PASSWORD", "blackspaceinc", "password of authentication client")
	fs.String("AUTHN_ISSUER_BASE_URL", "http://localhost", "authentication service issuer")
	fs.String("AUTHN_ORIGIN", "http://localhost", "origin of auth requests")
	fs.String("AUTHN_DOMAINS", "localhost", "authentication service domains")
	fs.String("AUTHN_PRIVATE_BASE_URL", "http://authentication_service",
		"authentication service private url. should be local host if these are not running on docker containers. "+
			"However if running in docker container with a configured docker network, the url should be equal to the service name")
	fs.String("AUTHN_PUBLIC_BASE_URL", "http://localhost", "authentication service public endpoint")
	fs.String("AUTHN_INTERNAL_PORT", "3000", "authentication service port")
	fs.String("AUTHN_PORT", "8404", "authentication service external port")
	fs.Bool("ENABLE_AUTH_SERVICE_PRIVATE_INTEGRATION", true, "enables communication with authentication service")
	// logging specific configurations
	fs.String("SERVICE_NAME", "authentication_handler_service", "service name")
	fs.String("JAEGER_ENDPOINT", "http://jaeger-collector:14268/api/traces", "jaeger collector endpoint")

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
	viper.SetDefault("ui-logo", "https://raw.githubusercontent.com/github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/gh-pages/cuddle_clap.gif")
	viper.Set("hostname", hostname)
	viper.Set("version", version.VERSION)
	viper.Set("revision", version.REVISION)
	viper.SetEnvPrefix("authentication_handler_service")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// load config from file
	if _, err := os.Stat(filepath.Join(viper.GetString("config-path"), viper.GetString("config"))); err == nil {
		viper.SetConfigName(strings.Split(viper.GetString("config"), ".")[0])
		viper.AddConfigPath(viper.GetString("config-path"))
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config file, %v\n", err)
		}
	}

	serviceName := viper.GetString("SERVICE_NAME")
	collectorEndpoint := viper.GetString("JAEGER_ENDPOINT")

	// initiaize a tracing object globally
	tracerEngine, closer := core_tracing.NewTracer(serviceName, collectorEndpoint, prometheus.New())
	defer closer.Close()

	if tracerEngine == nil {
		panic("cannot initialize tracer engine")
	}

	opentracing.SetGlobalTracer(tracerEngine.Tracer)

	coreMetrics := core_metrics.NewCoreMetricsEngineInstance(serviceName, nil)
	serviceMetrics := metrics.NewMetricsEngine(coreMetrics)

	ctx := context.Background()
	rootSpan := opentracing.SpanFromContext(ctx)

	// create logging object
	logger := core_logging.NewJSONLogger(nil, rootSpan)

	authnServiceClient := NewAuthServiceClient(err, logger)
	logger.InfoM("successfully initialized authentication service client")

	// start stress tests if any
	beginStressTest(viper.GetInt("STRESS_CPU"), viper.GetInt("STRESS_MEMORY"), logger)

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
		err := errors.New("`--random-delay-max` should be greater than `--random-delay-min`")
		logger.FatalM(err, "please fix configurations")
	}

	switch delayUnit := viper.GetString("random-delay-unit"); delayUnit {
	case
		"s",
		"ms":
		break
	default:
		err := errors.New("random-delay-unit` accepted values are: s|ms")
		logger.FatalM(err, "please fix configurations")
	}

	// load gRPC server config
	var grpcCfg grpc.Config
	if err := viper.Unmarshal(&grpcCfg); err != nil {
		err := errors.New("config unmarshal failed")
		logger.FatalM(err, "please fix configurations")
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

	// log version and revisions
	logger.InfoM("Starting authentication_handler_service",
		zap.String("version", viper.GetString("version")),
		zap.String("revision", viper.GetString("revision")),
		zap.String("port", srvCfg.Port))

	// start HTTP server
	srv, _ := api.NewServer(&srvCfg, authnServiceClient, logger, serviceMetrics.MicroServiceMetrics, serviceMetrics.Engine, tracerEngine)
	stopCh := signals.SetupSignalHandler()
	srv.ListenAndServe(stopCh)
}

func NewAuthServiceClient(err error, logger core_logging.ILog) *core_auth_sdk.Client {
	// initialize authentication client in order to establish communication with the
	// authentication service. This serves as a singular source of truth for authentication needs
	authUsername := viper.GetString("AUTHN_USERNAME")
	authPassword := viper.GetString("AUTHN_PASSWORD")
	domains := viper.GetString("AUTHN_DOMAINS")
	privateURL := viper.GetString("AUTHN_PRIVATE_BASE_URL") + ":" + viper.GetString("AUTHN_INTERNAL_PORT")
	origin := viper.GetString("AUTHN_ORIGIN")
	issuer := viper.GetString("AUTHN_ISSUER_BASE_URL") + ":" + viper.GetString("AUTHN_PORT")

	authnClient, err := initAuthnClient(authUsername, authPassword, domains, issuer, privateURL, origin)
	// crash the process if we cannot connect to the authentication service
	if err != nil {
		logger.FatalM(err, "failed to initialized authentication service client")
	}

	// TODO: make this a retryable operation
	retries := 1
	for retries < 4 {
		// perform a test request to the authentication service
		_, err = authnClient.ServerStats()
		if err != nil {
			if retries != 4 {
				logger.ErrorM(err, "failed to connect to authentication service")
			} else {
				logger.FatalM(err, "failed to connect to authentication service")
			}
			retries += 1
		} else {
			retries = 4
		}

		time.Sleep(1 * time.Second)
	}

	// attempt to connect to the authentication service if not then crash process
	return authnClient
}

func initZap(logLevel string) (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	switch logLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "fatal":
		level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case "panic":
		level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	}

	zapEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zapEncoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig.Build()
}

var stressMemoryPayload []byte

func beginStressTest(cpus int, mem int, logger core_logging.ILog) {
	done := make(chan int)
	if cpus > 0 {
		logger.InfoM("starting CPU stress", zap.Any("cores", cpus))
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
		path := "/tmp/authentication_handler_service.data"
		f, err := os.Create(path)

		if err != nil {
			logger.Error(err, "memory stress failed", "error")
		}

		if err := f.Truncate(1000000 * int64(mem)); err != nil {
			logger.Error(err, "memory stress failed", "error")
		}

		stressMemoryPayload, err = ioutil.ReadFile(path)
		f.Close()
		os.Remove(path)
		if err != nil {
			logger.Error(err, "memory stress failed", "error")
		}
		logger.InfoM("starting CPU stress", zap.Any("memory", len(stressMemoryPayload)))
	}
}

// initAuthnClient initializes an instance of the authn client primarily useful in
// communicating with the authentication service securely
func initAuthnClient(username, password, audience, issuer, url, origin string) (*core_auth_sdk.Client, error) {
	// Authentication.
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
