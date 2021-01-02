package utils

import (
	"context"
	"net/http"

	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	core_tracing "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/zap"
)

func StartRootSpanFromRequest(r *http.Request, operationType string, tracerEngine *core_tracing.TracingEngine, logger core_logging.ILog) (context.Context,
	opentracing.Span) {
	ctx := r.Context()
	logger.For(ctx).InfoM("HTTP request received", zap.String("method", r.Method), zap.Stringer("url", r.URL))

	// start a parent span
	spanCtx, _ := tracerEngine.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	parentSpan := tracerEngine.Tracer.StartSpan(operationType, ext.RPCServerOption(spanCtx))
	ctx = opentracing.ContextWithSpan(ctx, parentSpan)
	return ctx, parentSpan
}

func StartRootOperationSpan(ctx context.Context, operationType string, tracerEngine *core_tracing.TracingEngine, logger core_logging.ILog) (context.Context,
	opentracing.Span) {
	logger.For(ctx).InfoM("Starting parent span for operation")
	span, ctx := opentracing.StartSpanFromContext(ctx, operationType)
	ctx = opentracing.ContextWithSpan(ctx, span)
	return ctx, span
}
