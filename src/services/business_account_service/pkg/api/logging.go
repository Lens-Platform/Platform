package api

import (
	"net/http"

	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	"go.uber.org/zap"
)

type LoggingMiddleware struct {
	logger core_logging.ILog
}

func NewLoggingMiddleware(logger core_logging.ILog) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.InfoM(
			"request started",
			zap.String("model", r.Proto),
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("remote", r.RemoteAddr),
			zap.String("user-agent", r.UserAgent()),
		)
		next.ServeHTTP(w, r)
	})
}
