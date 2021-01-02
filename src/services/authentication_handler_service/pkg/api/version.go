package api

import (
	"net/http"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version"
)

// Version godoc
// @Summary Version
// @Description returns authentication_handler_service version and git commit hash
// @Tags HTTP API
// @Produce json
// @Router /version [get]
// @Success 200 {object} api.MapResponse
func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	result := map[string]string{
		"version": version.VERSION,
		"commit":  version.REVISION,
	}
	s.JSONResponse(w, r, result)
}
