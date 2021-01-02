package api

import (
	"net/http"
	"strconv"
	"time"

	utils "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-utilities"
	"github.com/opentracing/opentracing-go"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/constants"
)

// LockAccountResponse is struct providing errors tied to lock account operations
type LockAccountResponse struct {
	Error error `json:"error"`
}

// Lock account request
// swagger:parameters lockAccount
type LockAccountRequest struct {
	// id of the account to lock
	// in: query
	// required: true
	Id uint32 `json:"result"`
}

// swagger:route POST /v1/account/lock/{id} lockAccount
// Lock Account
//
// Locks an account through the authentication service
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https, ws, wss
//
//     Security:
//       api_key:
//       oauth: read, write
// responses:
//      200: operationResponse
// 400: badRequestError
// 404: notFoundError
// 403: forbiddenError
// 406: genericError
// 401: unAuthorizedError
// 500: internalServerError
// locks an account by account id
func (s *Server) lockAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx, parentSpan := s.startRootSpan(r, constants.LOCK_ACCOUNT)
	defer parentSpan.Finish()

	if s.IsNotAuthenticated(w, r) {
		return
	}

	var lockAccountResp LockAccountResponse
	ctx = opentracing.ContextWithSpan(ctx, parentSpan)
	// we extract the user id from the url initially
	authnID, err := s.ExtractIdOperationAndInstrument(ctx, r, constants.LOCK_ACCOUNT)
	if utils.HandleError(w, err, http.StatusInternalServerError) {
		s.logger.For(ctx).Error(err, "failed to parse account id from url")
		return
	}

	var (
		begin = time.Now()
		took  = time.Since(begin)
		f     = func() error {
			return s.authnClient.LockAccount(strconv.Itoa(int(authnID)))
		}
	)

	// TODO: perform this operation in a circuit breaker
	if err = s.RemoteOperationAndInstrument(ctx, f, constants.LOCK_ACCOUNT, &took); utils.HandleError(w, err,
		http.StatusInternalServerError) == true {
		s.logger.For(ctx).Error(err, "failed to lock created account")
		return
	}

	lockAccountResp.Error = err
	s.JSONResponse(w, r, lockAccountResp)
}
