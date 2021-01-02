package api

import (
	"net/http"

	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
)

type LoggingMiddleware struct {
}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		core_logging.JSONLogger.Info(
			"request started",
			"model", r.Proto,
			"uri", r.RequestURI,
			"method", r.Method,
			"remote", r.RemoteAddr,
			"user-agent", r.UserAgent(),
		)
		next.ServeHTTP(w, r)
	})
}
