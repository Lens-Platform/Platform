package api

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	core_auth_sdk "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-auth-sdk"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	middleware "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-middleware"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	gql "github.com/99designs/gqlgen/graphql/handler"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/database"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/fscache"
	graphql "github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/graphql_api/generated"
)

// @title business_account_service API
// @version 2.0
// @description Go microservice template for Kubernetes.

// @contact.name Source Code
// @contact.url https://github.com/stefanprodan/business_account_service

// @license.name MIT License
// @license.url https://github.com/stefanprodan/business_account_service/blob/master/LICENSE

// @host localhost:9898
// @BasePath /
// @schemes http https

var (
	healthy int32
	ready   int32
	watcher *fscache.Watcher
)

type Config struct {
	HttpClientTimeout         time.Duration `mapstructure:"http-client-timeout"`
	HttpServerTimeout         time.Duration `mapstructure:"http-server-timeout"`
	HttpServerShutdownTimeout time.Duration `mapstructure:"http-server-shutdown-timeout"`
	BackendURL                []string      `mapstructure:"backend-url"`
	UILogo                    string        `mapstructure:"ui-logo"`
	UIMessage                 string        `mapstructure:"ui-message"`
	UIColor                   string        `mapstructure:"ui-color"`
	UIPath                    string        `mapstructure:"ui-path"`
	DataPath                  string        `mapstructure:"data-path"`
	ConfigPath                string        `mapstructure:"config-path"`
	CertPath                  string        `mapstructure:"cert-path"`
	Port                      string        `mapstructure:"port"`
	SecurePort                string        `mapstructure:"secure-port"`
	PortMetrics               int           `mapstructure:"port-metrics"`
	Hostname                  string        `mapstructure:"hostname"`
	H2C                       bool          `mapstructure:"h2c"`
	RandomDelay               bool          `mapstructure:"random-delay"`
	RandomDelayUnit           string        `mapstructure:"random-delay-unit"`
	RandomDelayMin            int           `mapstructure:"random-delay-min"`
	RandomDelayMax            int           `mapstructure:"random-delay-max"`
	RandomError               bool          `mapstructure:"random-error"`
	Unhealthy                 bool          `mapstructure:"unhealthy"`
	Unready                   bool          `mapstructure:"unready"`
	JWTSecret                 string        `mapstructure:"jwt-secret"`
	CacheServer               string        `mapstructure:"cache-server"`
}

type Server struct {
	router        *mux.Router
	logger        core_logging.ILog
	tracingEngine *core_tracing.TracingEngine
	metricsEngine *core_metrics.CoreMetricsEngine
	config        *Config
	pool          *redis.Pool
	handler       http.Handler
	gqlServer     *gql.Server
	db            *database.Db
	authClient    *core_auth_sdk.Client
}

func NewServer(config *Config, logger core_logging.ILog, tracer *core_tracing.TracingEngine, metrics *core_metrics.CoreMetricsEngine,
	db *database.Db, authClient *core_auth_sdk.Client) (*Server, error) {
	gqlServer := gql.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graphql.Resolver{
		Db:      db,
		Logger:  logger,
		Tracer:  tracer,
		Metrics: metrics,
	}}))

	srv := &Server{
		router:        mux.NewRouter(),
		logger:        logger,
		config:        config,
		gqlServer:     gqlServer,
		db:            db,
		tracingEngine: tracer,
		metricsEngine: metrics,
		authClient:    authClient,
	}

	return srv, nil
}

func (s *Server) registerHandlers() {
	s.router.Handle("/metrics", promhttp.Handler())

	s.router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	s.router.HandleFunc("/", s.indexHandler).HeadersRegexp("User-Agent", "^Mozilla.*").Methods("GET")
	s.router.HandleFunc("/", s.infoHandler).Methods("GET")
	s.router.HandleFunc("/version", s.versionHandler).Methods("GET")
	s.router.HandleFunc("/env", s.envHandler).Methods("GET", "POST")
	s.router.HandleFunc("/delay/{wait:[0-9]+}", s.delayHandler).Methods("GET").Name("delay")
	s.router.HandleFunc("/healthz", s.healthzHandler).Methods("GET")
	s.router.HandleFunc("/readyz", s.readyzHandler).Methods("GET")
	s.router.HandleFunc("/readyz/enable", s.enableReadyHandler).Methods("POST")
	s.router.HandleFunc("/readyz/disable", s.disableReadyHandler).Methods("POST")
	s.router.HandleFunc("/panic", s.panicHandler).Methods("GET")
	s.router.HandleFunc("/status/{code:[0-9]+}", s.statusHandler).Methods("GET", "POST", "PUT").Name("status")
	s.router.HandleFunc("/cache/{key}", s.cacheWriteHandler).Methods("POST", "PUT")
	s.router.HandleFunc("/cache/{key}", s.cacheDeleteHandler).Methods("DELETE")
	s.router.HandleFunc("/cache/{key}", s.cacheReadHandler).Methods("GET").Name("cache")
	s.router.HandleFunc("/configs", s.configReadHandler).Methods("GET")
	s.router.HandleFunc("/api/info", s.infoHandler).Methods("GET")
	s.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	s.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	s.router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		doc, err := swag.ReadDoc()
		if err != nil {
			s.logger.Error(errors.ErrSwaggerGenError, err.Error(), zap.Error(err), zap.String("path", "/swagger.json"))
		}
		w.Write([]byte(doc))
	})

	s.router.Handle("/graphql", playground.Handler("GraphQL playground", "/query"))
	s.router.Handle("/query", s.gqlServer)
}

func (s *Server) registerMiddlewares() {
	prom := NewPrometheusMiddleware()
	s.router.Use(prom.Handler)
	httpLogger := NewLoggingMiddleware(s.logger)
	s.router.Use(httpLogger.Handler)
	s.router.Use(versionMiddleware)

	authMw := middleware.NewAuthnMw(s.authClient, s.logger, "business_account_service")
	s.router.Use(authMw.AuthenticationMiddleware)

	if s.config.RandomDelay {
		randomDelayer := NewRandomDelayMiddleware(s.config.RandomDelayMin, s.config.RandomDelayMax, s.config.RandomDelayUnit)
		s.router.Use(randomDelayer.Handler)
	}
	if s.config.RandomError {
		s.router.Use(randomErrorMiddleware)
	}
}

func (s *Server) ListenAndServe(stopCh <-chan struct{}) {
	go s.startMetricsServer()

	s.registerHandlers()
	s.registerMiddlewares()

	if s.config.H2C {
		s.handler = h2c.NewHandler(s.router, &http2.Server{})
	} else {
		s.handler = s.router
	}

	// s.printRoutes()

	// load configs in memory and start watching for changes in the config dir
	if stat, err := os.Stat(s.config.ConfigPath); err == nil && stat.IsDir() {
		var err error
		watcher, err = fscache.NewWatch(s.config.ConfigPath)
		if err != nil {
			s.logger.Error(errors.ErrFailedToWatchConfigDirectory, "config watch error", zap.Error(err), zap.String("path", s.config.ConfigPath))
		} else {
			watcher.Watch()
		}
	}

	// start redis connection pool
	ticker := time.NewTicker(30 * time.Second)
	s.startCachePool(ticker, stopCh)

	// create the http server
	srv := s.startServer()

	// create the secure server
	secureSrv := s.startSecureServer()

	// signal Kubernetes the server is ready to receive traffic
	if !s.config.Unhealthy {
		atomic.StoreInt32(&healthy, 1)
	}
	if !s.config.Unready {
		atomic.StoreInt32(&ready, 1)
	}

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(context.Background(), s.config.HttpServerShutdownTimeout)
	defer cancel()

	// all calls to /healthz and /readyz will fail from now on
	atomic.StoreInt32(&healthy, 0)
	atomic.StoreInt32(&ready, 0)

	// close cache pool
	if s.pool != nil {
		_ = s.pool.Close()
	}

	s.logger.Info("Shutting down HTTP/HTTPS server", zap.Duration("timeout", s.config.HttpServerShutdownTimeout))

	// wait for Kubernetes readiness probe to remove this instance from the load balancer
	// the readiness check interval must be lower than the timeout
	if viper.GetString("level") != "debug" {
		time.Sleep(3 * time.Second)
	}

	// determine if the http server was started
	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Error(errors.ErrHttpServerFailedGracefuleShutdown, err.Error())
		}
	}

	// determine if the secure server was started
	if secureSrv != nil {
		if err := secureSrv.Shutdown(ctx); err != nil {
			s.logger.Error(errors.ErrHttpsServerFailedGracefuleShutdown, err.Error())
		}
	}
}

func (s *Server) startServer() *http.Server {

	// determine if the port is specified
	if s.config.Port == "0" {

		// move on immediately
		return nil
	}

	srv := &http.Server{
		Addr:         ":" + s.config.Port,
		WriteTimeout: s.config.HttpServerTimeout,
		ReadTimeout:  s.config.HttpServerTimeout,
		IdleTimeout:  2 * s.config.HttpServerTimeout,
		Handler:      s.handler,
	}

	// start the server in the background
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.FatalM(errors.ErrHttpServerCrashed, err.Error())
		}
	}()

	// return the server and routine
	return srv
}

func (s *Server) startSecureServer() *http.Server {

	// determine if the port is specified
	if s.config.SecurePort == "0" {

		// move on immediately
		return nil
	}

	srv := &http.Server{
		Addr:         ":" + s.config.SecurePort,
		WriteTimeout: s.config.HttpServerTimeout,
		ReadTimeout:  s.config.HttpServerTimeout,
		IdleTimeout:  2 * s.config.HttpServerTimeout,
		Handler:      s.handler,
	}

	cert := path.Join(s.config.CertPath, "tls.crt")
	key := path.Join(s.config.CertPath, "tls.key")

	// start the server in the background
	go func() {
		if err := srv.ListenAndServeTLS(cert, key); err != http.ErrServerClosed {
			s.logger.FatalM(errors.ErrHttpsServerCrashed, err.Error())
		}
	}()

	// return the server
	return srv
}

func (s *Server) startMetricsServer() {
	if s.config.PortMetrics > 0 {
		mux := http.DefaultServeMux
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%v", s.config.PortMetrics),
			Handler: mux,
		}

		srv.ListenAndServe()
	}
}

func (s *Server) printRoutes() {
	s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
}

type ArrayResponse []string
type MapResponse map[string]string
