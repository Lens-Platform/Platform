package api

import (
	"net/http"
	"strconv"
	"time"

	utils "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-utilities"
	"github.com/opentracing/opentracing-go"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/constants"
)

// UnLockAccountResponse is struct providing errors tied to Unlock account operations
type UnLockAccountResponse struct {
	Error error `json:"error"`
}

// UnLock account request
// swagger:parameters unlockAccount
type UnLockAccountRequest struct {
	// id of the account to unlock
	// in: query
	// required: true
	Id uint32 `json:"result"`
}

// swagger:route POST /v1/account/unlock/{id} unlockAccount
// UnLock Account
//
// UnLocks an account through the authentication service
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
// unlocks an by account id
func (s *Server) unlockAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx, parentSpan := s.startRootSpan(r, constants.UNLOCK_ACCOUNT)
	defer parentSpan.Finish()

	if s.IsNotAuthenticated(w, r) {
		return
	}

	var unlockAccountResp UnLockAccountResponse

	ctx = opentracing.ContextWithSpan(ctx, parentSpan)
	authnID, err := s.ExtractIdOperationAndInstrument(ctx, r, constants.UNLOCK_ACCOUNT)
	if utils.HandleError(w, err, http.StatusInternalServerError) {
		s.logger.For(ctx).Error(err, "failed to parse account id from url")
		return
	}

	var (
		begin = time.Now()
		took  = time.Since(begin)
		f     = func() error {
			return s.authnClient.UnlockAccount(strconv.Itoa(int(authnID)))
		}
	)

	// TODO: perform this operation in a circuit breaker
	if err = s.RemoteOperationAndInstrument(ctx, f, constants.UNLOCK_ACCOUNT, &took); utils.HandleError(w, err,
		http.StatusInternalServerError) == true {
		s.logger.For(ctx).Error(err, "failed to unlock created account")
		return
	}

	unlockAccountResp.Error = err
	s.JSONResponse(w, r, unlockAccountResp)
}
