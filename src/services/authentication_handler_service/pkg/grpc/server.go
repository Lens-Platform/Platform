package grpc

import (
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
)

type Server struct {
	config *Config
	logger core_logging.ILog
}

type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-service-name"`
}

func NewServer(config *Config, logger core_logging.ILog) (*Server, error) {
	srv := &Server{
		config: config,
		logger: logger,
	}

	return srv, nil
}

func (s *Server) ListenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.config.Port))
	if err != nil {
		s.logger.FatalM(errors.New(fmt.Sprintf("failed to listen on port: %d", s.config.Port)), "failed to listen")
	}

	srv := grpc.NewServer()
	server := health.NewServer()
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, server)
	server.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	if err := srv.Serve(listener); err != nil {
		s.logger.FatalM(err, "failed to serve")
	}
}
