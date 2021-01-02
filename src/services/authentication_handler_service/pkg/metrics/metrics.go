package metrics

import (
	"fmt"

	core_metrics "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// TODO: update this hard coded service name
var ServiceName string = "authentication_handler_service"

type MetricsEngine struct {
	MicroServiceMetrics *CoreMetrics
	Engine              *core_metrics.CoreMetricsEngine
}

func NewMetricsEngine(engine *core_metrics.CoreMetricsEngine) *MetricsEngine {
	return &MetricsEngine{
		MicroServiceMetrics: NewCoreMetrics(engine),
		Engine:              engine,
	}
}

type CoreMetrics struct {
	// tracks the number of http requests partitioned by name and status code
	// used for monitoring and alerting (RED method)
	HttpRequestCounter *core_metrics.CounterVec
	// tracks the latencies associated with a http requests by operation name
	// used for horizontal pod auto-scaling (Kubernetes HPA v2)
	HttpRequestLatencyCounter *core_metrics.HistogramVec
	// tracks the number of times there was a failure or success when trying to extract id from the request url
	ExtractIdOperationCounter *core_metrics.CounterVec
	// tracks the number of times there was a failure or success when trying to extract id from the request url
	RemoteOperationStatusCounter    *core_metrics.CounterVec
	RemoteOperationsLatencyCounter  *core_metrics.HistogramVec
	InvalidRequestParametersCounter *core_metrics.CounterVec
	CastingOperationFailureCounter  *core_metrics.CounterVec
	DecodeRequestStatusCounter      *core_metrics.CounterVec
}

func NewCoreMetrics(engine *core_metrics.CoreMetricsEngine) *CoreMetrics {
	return &CoreMetrics{
		HttpRequestCounter:              NewHttpRequestCounter(engine),
		HttpRequestLatencyCounter:       NewHttpRequestLatencyCounter(engine),
		ExtractIdOperationCounter:       NewExtractIdOperationCounter(engine),
		RemoteOperationStatusCounter:    NewRemoteOperationStatusCounter(engine),
		RemoteOperationsLatencyCounter:  NewRemoteOperationLatencyCounter(engine),
		InvalidRequestParametersCounter: NewInvalidRequestParametersCounter(engine),
		CastingOperationFailureCounter:  NewCastingOperationFailureCounter(engine),
		DecodeRequestStatusCounter:      NewDecodeRequestStatusCounter(engine),
	}
}

func NewHttpRequestCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.CounterVec {
	newCounter := core_metrics.NewCounterVec(&core_metrics.CounterOpts{
		Namespace: ServiceName,
		Subsystem: "HTTP",
		Name:      fmt.Sprintf("%s_http_requests_total", ServiceName),
		Help:      "How many HTTP requests processed partitioned by name and status code",
	}, []string{"name", "code"})

	engine.RegisterMetric(newCounter)
	return newCounter
}

func NewHttpRequestLatencyCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.HistogramVec {
	newCounter := core_metrics.NewHistogramVec(&core_metrics.HistogramOpts{
		Namespace:         ServiceName,
		Subsystem:         "HTTP",
		Name:              fmt.Sprintf("%s_http_requests_latencies", ServiceName),
		Help:              "Seconds spent serving HTTP requests.",
		ConstLabels:       nil,
		Buckets:           prometheus.DefBuckets,
		DeprecatedVersion: "",
		StabilityLevel:    "",
	}, []string{"method", "path", "status"})
	engine.RegisterMetric(newCounter)
	return newCounter
}

func NewExtractIdOperationCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.CounterVec {
	// tracks the number of times there was a failure or success when trying to extract id from the request url
	newCounter := core_metrics.NewCounterVec(&core_metrics.CounterOpts{
		Namespace: ServiceName,
		Subsystem: "HTTP",
		Name:      fmt.Sprintf("%s_status_of_extract_id_operation_from_requests_total", ServiceName),
		Help:      "The status of the extract the id operation from the HTTP requests processed partitioned by operation name and operation status",
	}, []string{"operation_name", "status"})
	engine.RegisterMetric(newCounter)
	return newCounter
}

func NewRemoteOperationStatusCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.CounterVec {
	newCounter := core_metrics.NewCounterVec(&core_metrics.CounterOpts{
		Namespace: ServiceName,
		Subsystem: "HTTP",
		Name:      fmt.Sprintf("%s_status_of_remote_operation_total", ServiceName),
		Help:      "A count of the status all remote operations operation",
	}, []string{"operation_name", "status"})
	engine.RegisterMetric(newCounter)
	return newCounter
}

func NewRemoteOperationLatencyCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.HistogramVec {
	newCounter := core_metrics.NewHistogramVec(&core_metrics.HistogramOpts{
		Namespace:         ServiceName,
		Subsystem:         "HTTP",
		Name:              fmt.Sprintf("%s_remote_operation_requests_latencies", ServiceName),
		Help:              "Seconds spent serving remote operations HTTP requests.",
		ConstLabels:       nil,
		Buckets:           prometheus.DefBuckets,
		DeprecatedVersion: "",
		StabilityLevel:    "",
	}, []string{"operation", "status"})
	engine.RegisterMetric(newCounter)
	return newCounter
}

func NewInvalidRequestParametersCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.CounterVec {
	newCounter := core_metrics.NewCounterVec(&core_metrics.CounterOpts{
		Namespace: ServiceName,
		Subsystem: "HTTP",
		Name:      fmt.Sprintf("%s_invalid_request_parameters_total", ServiceName),
		Help:      "A count of the total number of invalid request parameter count",
	}, []string{"operation_name"})
	engine.RegisterMetric(newCounter)
	return newCounter
}

func NewCastingOperationFailureCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.CounterVec {
	newCounter := core_metrics.NewCounterVec(&core_metrics.CounterOpts{
		Namespace: ServiceName,
		Subsystem: "HTTP",
		Name:      fmt.Sprintf("%s_casting_operation_failure_total", ServiceName),
		Help:      "A count of the total number of failed casts from interface to object",
	}, []string{"operation_name"})
	engine.RegisterMetric(newCounter)
	return newCounter
}

func NewDecodeRequestStatusCounter(engine *core_metrics.CoreMetricsEngine) *core_metrics.CounterVec {
	newCounter := core_metrics.NewCounterVec(&core_metrics.CounterOpts{
		Namespace: ServiceName,
		Subsystem: "HTTP",
		Name:      fmt.Sprintf("%s_decoder_request_op_counter_total", ServiceName),
		Help:      "A count of the status of all decode operations",
	}, []string{"operation_name", "status"})
	engine.RegisterMetric(newCounter)
	return newCounter
}
