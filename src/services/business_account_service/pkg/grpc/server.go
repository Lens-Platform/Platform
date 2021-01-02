package grpc

import (
	"fmt"
	"net"

	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/errors"
)

type Server struct {
	logger        core_logging.ILog
	config        *Config
	tracingEngine *core_tracing.TracingEngine
	metricsEngine *core_metrics.CoreMetricsEngine
}

type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-service-name"`
}

func NewServer(config *Config, logger core_logging.ILog, tracer *core_tracing.TracingEngine, metrics *core_metrics.CoreMetricsEngine) (*Server, error) {
	srv := &Server{
		logger:        logger,
		config:        config,
		tracingEngine: tracer,
		metricsEngine: metrics,
	}

	return srv, nil
}

func (s *Server) ListenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.config.Port))
	if err != nil {
		s.logger.FatalM(errors.ErrFailedToStartGRPCServer, fmt.Sprintf("failed to listen on port %d", s.config.Port))
	}

	srv := grpc.NewServer()
	server := health.NewServer()
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, server)
	server.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	if err := srv.Serve(listener); err != nil {
		s.logger.FatalM(errors.ErrFailedToStartGRPCServer, err.Error())
	}
}
