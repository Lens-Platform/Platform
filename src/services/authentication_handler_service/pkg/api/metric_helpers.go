package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/constants"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/helper"
)

// ExtractIdOperationAndInstrument extracts an account id from a request and increments the necessary metrics
func (s *Server) ExtractIdOperationAndInstrument(ctx context.Context, r *http.Request, operation string) (uint32, error) {
	// we start a child span for the rpc operation
	extractIdChildSpan := s.tracerEngine.CreateChildSpan(ctx, "EXTRACT_ID_FROM_REQUEST_HEADER")
	defer extractIdChildSpan.Finish()

	var status = constants.SUCCESS
	authnID, err := helper.ExtractIDFromRequest(r)
	if err != nil {
		status = constants.FAILURE
	}

	s.metrics.ExtractIdOperationCounter.WithLabelValues(operation, status).Inc()
	return authnID, err
}

func (s *Server) RemoteOperationAndInstrument(ctx context.Context, f func() error, operationType string, took *time.Duration) error {
	// we start a child span for the rpc operation
	authnSvcRpcSpan := s.tracerEngine.CreateChildSpan(ctx, fmt.Sprintf("ATHENTICATION_SERVICE_%s_RPC_REQUEST", operationType))
	defer authnSvcRpcSpan.Finish()

	var status = constants.SUCCESS
	err := f()
	if err != nil {
		status = constants.FAILURE
	}

	s.metrics.RemoteOperationStatusCounter.WithLabelValues(operationType, status).Inc()
	s.metrics.RemoteOperationsLatencyCounter.WithLabelValues(operationType, status).Observe(took.Seconds())
	return err
}

func (s *Server) RemoteOperationAndInstrumentWithResult(
	ctx context.Context,
	f func() (interface{}, error),
	operationType string,
	took *time.Duration) (interface{}, error) {

	authnSvcRpcSpan := s.tracerEngine.CreateChildSpan(ctx, fmt.Sprintf("ATHENTICATION_SERVICE_%s_RPC_REQUEST", operationType))
	defer authnSvcRpcSpan.Finish()

	var status = constants.SUCCESS
	result, err := f()
	if err != nil {
		status = constants.FAILURE
	}

	s.metrics.RemoteOperationStatusCounter.WithLabelValues(operationType, status).Inc()
	s.metrics.RemoteOperationsLatencyCounter.WithLabelValues(operationType, status).Observe(took.Seconds())
	return result, err
}

func (s *Server) DecodeRequestAndInstrument(ctx context.Context, w http.ResponseWriter, r *http.Request, obj interface{},
	operationType string) error {
	// we start a child span for the rpc operation
	childSpan := s.tracerEngine.CreateChildSpan(ctx, "DECODE_REQUEST")
	defer childSpan.Finish()

	var status = constants.SUCCESS

	err := helper.DecodeJSONBody(w, r, &obj)
	if err != nil {
		status = constants.FAILURE
	}

	s.metrics.DecodeRequestStatusCounter.WithLabelValues(operationType, status).Inc()
	return err
}
