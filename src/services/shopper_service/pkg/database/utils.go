package database

import (
	"context"
	"io"
	"net/http"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/errors"
	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/utils"
	"github.com/giantswarm/retry-go"
	"github.com/opentracing/opentracing-go"
)

// performHttpRequest performs an http request and returns the request response status and any occuring errors
func (db *Db) performHttpRequest(httpClient *http.Client, httpReq *http.Request, ctx context.Context) (*http.Response, error) {
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		db.Logger.For(ctx).Error(errors.ErrDistributedTransactionError, err.Error())
		return nil, err
	}
	return resp, nil
}

// startRootSpan starts a root span object
func (db *Db) startRootSpan(ctx context.Context, operationType string) (context.Context, opentracing.Span) {
	return utils.StartRootOperationSpan(ctx, operationType, db.TracingEngine, db.Logger)
}

// processResponseStatusCode processes response code and return wether the status code was faulty or not
func (db *Db) processResponseStatusCode(resp *http.Response, ctx context.Context) bool {
	if resp.StatusCode != http.StatusOK {
		db.Logger.For(ctx).Error(errors.ErrDistributedTransactionError, errors.ErrDistributedTransactionError.Error()+" authentication handler service")
		return true
	}
	return false
}

// createRequestAndPropagateTraces creates a request and propagates the contexts to tracing engine
func (db *Db) createRequestAndPropagateTraces(url string, childSpan opentracing.Span, body io.Reader) *http.Request {
	httpReq, _ := http.NewRequest("POST", url, body)

	// Transmit the span's TraceContext as HTTP headers on our
	// outbound request.
	opentracing.GlobalTracer().Inject(
		childSpan.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(httpReq.Header))
	return httpReq
}

func (db *Db) PerformRetryableOperation(f func() error) error {
	return retry.Do(func() error {
		return f()
	},
		retry.MaxTries(db.MaxRetriesPerOperation),
		// TODO: emit metrics
		// retry.AfterRetryLimit()
		retry.Timeout(db.RetryTimeOut),
		retry.Sleep(db.OperationSleepInterval),
	)
}
