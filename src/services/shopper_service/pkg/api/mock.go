package api

import (
	"time"

	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	"github.com/gorilla/mux"
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

	return &Server{
		router: mux.NewRouter(),
		logger: core_logging.JSONLogger,
		config: config,
	}
}
