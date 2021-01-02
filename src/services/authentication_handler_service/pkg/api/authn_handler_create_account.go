package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	utils "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-utilities"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/constants"
)

type CreateAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountResponse struct {
	Error error  `json:"error"`
	Id    uint32 `json:"id"`
}

// Create account request
// swagger:parameters createAccountRequest
type createAccountRequest struct {
	// in: body
	Body struct {
		// user username to create
		// required : true
		Email string `json:"email"`
		// user password to create
		// required : true
		Password string `json:"password"`
	}
}

// Account successfully created
// swagger:response createAccountResponse
type createAccountResponse struct {
	// in: body
	Body struct {
		// account id
		// required: true
		// example: 20
		Id uint32 `json:"id"`
		// error
		// required: true
		// example: account already exists
		Error error `json:"error"`
	}
}

// swagger:route POST /v1/account/create AccountDetails createAccountRequest
//
// Create Account
//
// creates an account object via the authentication service
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https, ws, wss
//
//
//     Security:
//       api_key:
//       oauth: read, write
// responses:
//      200: signUpUserResp
// 400: badRequestError
// 404: notFoundError
// 403: forbiddenError
// 406: genericError
// 500: internalServerError
// creates an account
func (s *Server) createAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx, parentSpan := s.startRootSpan(r, constants.CREATE_ACCOUNT)
	defer parentSpan.Finish()

	var (
		createAccountReq CreateAccountRequest
	)

	err := s.DecodeRequestAndInstrument(ctx, w, r, &createAccountReq, constants.CREATE_ACCOUNT)
	if utils.HandleError(w, err, http.StatusInternalServerError) {
		s.logger.For(ctx).ErrorM(err, "failed to decode request")
		return
	}

	if createAccountReq.Password == "" || createAccountReq.Email == "" {
		s.metrics.InvalidRequestParametersCounter.WithLabelValues(constants.CREATE_ACCOUNT).Inc()
		errMsg := "invalid input parameters. please specify a username and password"
		s.logger.For(ctx).ErrorM(err, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	var (
		begin = time.Now()
		took  = time.Since(begin)
		f     = func() (interface{}, error) {
			return s.authnClient.ImportAccount(createAccountReq.Email, createAccountReq.Password, false)
		}
	)

	ctx = opentracing.ContextWithSpan(ctx, parentSpan)
	result, err := s.RemoteOperationAndInstrumentWithResult(ctx, f, constants.CREATE_ACCOUNT, &took)
	if utils.HandleError(w, err, http.StatusInternalServerError) {
		s.logger.For(ctx).ErrorM(err, "failed to create account via authentication service")
		return
	}

	authnID, ok := result.(int)
	if !ok {
		s.metrics.CastingOperationFailureCounter.WithLabelValues(constants.CREATE_ACCOUNT)
		err := errors.New("failed to convert result to uint32 id value")
		s.logger.For(ctx).ErrorM(err, "casting error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// this is ran in the case any error is encountered when this function returns. We don't want to leave the datastore of the authentication
	// service in an inconsistent state
	defer s.ArchiveAccountIfErrorsOccur(ctx, err, authnID, parentSpan)

	s.logger.For(ctx).Info("Successfully created user account", zap.Int("accountID", int(authnID)))
	response := CreateAccountResponse{Id: uint32(authnID), Error: err}
	s.JSONResponse(w, r, response)
}

func (s *Server) ArchiveAccountIfErrorsOccur(ctx context.Context, err error, authnID int, parentSpan opentracing.Span) {
	func() {
		if err != nil {
			startTime := time.Now()
			elapsedTime := time.Since(startTime)
			op := func() error {
				return s.authnClient.ArchiveAccount(strconv.Itoa(int(authnID)))
			}

			// we start a child span for the rpc operation
			ctx = opentracing.ContextWithSpan(ctx, parentSpan)

			// TODO: perform this operation in a circuit breaker
			s.logger.ErrorM(err, "unable to create user account in authentication service. archiving account")
			if err = s.RemoteOperationAndInstrument(ctx, op, constants.DELETE_ACCOUNT, &elapsedTime); err != nil {
				s.logger.ErrorM(err, "failed to archive created account")
			}
		}
	}()
}
