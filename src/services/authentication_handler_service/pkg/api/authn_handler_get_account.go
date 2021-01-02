package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	core_auth_sdk "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-auth-sdk"
	utils "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-utilities"
	"github.com/opentracing/opentracing-go"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/constants"
)

// GetAccountResponse is struct providing errors tied to get account operations
type GetAccountResponse struct {
	Account *core_auth_sdk.Account `json:"account"`
	Error   error                  `json:"error"`
}

// Get account by id request
// swagger:parameters getAccount
type GetAccountRequest struct {
	// id of the account to obtain
	// in: query
	// required: true
	Id uint32 `json:"result"`
}

// swagger:route GET /v1/account/{id} getAccount
// Get Account
//
// Gets an account through the authentication service
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
// deletes an by account id
func (s *Server) getAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx, parentSpan := s.startRootSpan(r, constants.GET_ACCOUNT)
	defer parentSpan.Finish()

	if s.IsNotAuthenticated(w, r) {
		return
	}

	ctx = opentracing.ContextWithSpan(ctx, parentSpan)
	// we extract the user id from the url initially
	authnID, err := s.ExtractIdOperationAndInstrument(ctx, r, constants.GET_ACCOUNT)
	if err != nil {
		s.logger.For(ctx).ErrorM(err, "failed to parse account id from url")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		begin = time.Now()
		took  = time.Since(begin)
		f     = func() (interface{}, error) {
			return s.authnClient.GetAccount(strconv.Itoa(int(authnID)))
		}
	)

	result, err := s.RemoteOperationAndInstrumentWithResult(ctx, f, constants.GET_ACCOUNT, &took)
	if utils.HandleError(w, err, http.StatusInternalServerError) {
		s.logger.For(ctx).Error(err, "failed to get account")
		return
	}

	account, ok := result.(*core_auth_sdk.Account)
	if !ok {
		s.metrics.CastingOperationFailureCounter.WithLabelValues(constants.GET_ACCOUNT)
		err := errors.New("failed to cast response to account object")
		s.logger.For(ctx).ErrorM(err, "casting failure")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var getAccountResp = GetAccountResponse{
		Account: account,
		Error:   err,
	}

	s.JSONResponse(w, r, getAccountResp)
}
