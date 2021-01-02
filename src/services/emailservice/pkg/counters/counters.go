package counters

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/BlackspaceInc/Backend/common/counters/infrastructure"
)

// ServiceLevelCounters witholds references counters specific to the email services
type ServiceLevelCounters struct {
	NumEmailsSent                  prometheus.Counter
	NumEmailsReceived              prometheus.Counter
	NumMessageNacks                prometheus.Counter
	NumMessageAcks                 prometheus.Counter
	SendgridEmailLatency           prometheus.SummaryVec
	FailedSendgridRequestCount     prometheus.Counter
	SuccessfulSendgridRequestCount prometheus.Counter
	FailedEmailUnMarshallingEvents prometheus.Counter
}

type Telemetry struct {
	InfrastructureCounter *infrastructure.InfrastructureCounters
	ServiceCounters       *ServiceLevelCounters
}

// New initializes a set of counters for the email service
func New(servicename string) *ServiceLevelCounters {
	// https://www.alexedwards.net/blog/configuring-sqldb
	numEmailsSentCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicename,
			Subsystem: "service",
			Name:      "total_emails_sent",
			Help:      "Number of total emails sent",
		})

	numEmailsReceivedCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicename,
			Subsystem: "service",
			Name:      "total_emails_received",
			Help:      "Number of total emails received",
		})

	numMessageNacksCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicename,
			Subsystem: "service",
			Name:      "amqp_message_nacks",
			Help:      "Number of failed amqp message acks",
		})

	numMessageAcksCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicename,
			Name:      "amqp_message_acks",
			Help:      "Number of successfully amqp message acks",
		})

	sendgridLatencyCounter := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  servicename,
			Name:       "sendgrid_latency",
			Help:       "latency of sendgrid api requests",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"duration"},
	)

	numFailedSendgridRequestCount := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicename,
			Name:      "failed_sendgrid_request_count",
			Help:      "Number of failed sendgrid request count",
		})

	numSuccessfulSendgridRequestCount := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicename,
			Name:      "successful_sendgrid_request_count",
			Help:      "Number of successfully sendgrid request count",
		})

	numFailedUnmarshellingEmailEvents := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: servicename,
			Name:      "failed_to_unmarshall_email_from_queue",
			Help:      "Number of failed email unmarshelling events",
		})

	return &ServiceLevelCounters{
		NumEmailsSent:                  numEmailsSentCounter,
		NumEmailsReceived:              numEmailsReceivedCounter,
		NumMessageNacks:                numMessageNacksCounter,
		NumMessageAcks:                 numMessageAcksCounter,
		SendgridEmailLatency:           *sendgridLatencyCounter,
		FailedSendgridRequestCount:     numFailedSendgridRequestCount,
		SuccessfulSendgridRequestCount: numSuccessfulSendgridRequestCount,
		FailedEmailUnMarshallingEvents: numFailedUnmarshellingEmailEvents,
	}
}

// RegisterMetrics registers the service level counters to prometheus
func (rc *ServiceLevelCounters) RegisterMetrics() {
	_ = infrastructure.RegisterCounterMetric(rc.NumEmailsSent)
	_ = infrastructure.RegisterCounterMetric(rc.NumEmailsReceived)
	_ = infrastructure.RegisterCounterMetric(rc.NumMessageNacks)
	_ = infrastructure.RegisterCounterMetric(rc.NumMessageAcks)
	_ = infrastructure.RegisterSummaryVecMetric(rc.SendgridEmailLatency)
	_ = infrastructure.RegisterCounterMetric(rc.FailedSendgridRequestCount)
	_ = infrastructure.RegisterCounterMetric(rc.SuccessfulSendgridRequestCount)
	_ = infrastructure.RegisterCounterMetric(rc.FailedEmailUnMarshallingEvents)
}
