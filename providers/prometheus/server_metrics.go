// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package prometheus

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a gRPC server.
type ServerMetrics struct {
	ipLabelsEnabled         bool
	serverStartedCounter    *prometheus.CounterVec
	serverHandledCounter    *prometheus.CounterVec
	serverStreamMsgReceived *prometheus.CounterVec
	serverStreamMsgSent     *prometheus.CounterVec
	// serverHandledHistogram can be nil.
	serverHandledHistogram *prometheus.HistogramVec
}

// NewServerMetrics returns a new ServerMetrics object that has server interceptor methods.
// NOTE: Remember to register ServerMetrics object by using prometheus registry
// e.g. prometheus.MustRegister(myServerMetrics).
func NewServerMetrics(opts ...ServerMetricsOption) *ServerMetrics {
	var config serverMetricsConfig
	config.apply(opts)

	addIPLables := func(orig []string) []string {
		if config.ipLabelsEnabled {
			return serverMetricAddIPLabel(orig)
		}
		return orig
	}

	sm := &ServerMetrics{
		ipLabelsEnabled: config.ipLabelsEnabled,
		serverStartedCounter: prometheus.NewCounterVec(
			config.counterOpts.apply(prometheus.CounterOpts{
				Name: "grpc_server_started_total",
				Help: "Total number of RPCs started on the server.",
			}), addIPLables([]string{"grpc_type", "grpc_service", "grpc_method"})),
		serverHandledCounter: prometheus.NewCounterVec(
			config.counterOpts.apply(prometheus.CounterOpts{
				Name: "grpc_server_handled_total",
				Help: "Total number of RPCs completed on the server, regardless of success or failure.",
			}), addIPLables([]string{"grpc_type", "grpc_service", "grpc_method", "grpc_code"})),
		serverStreamMsgReceived: prometheus.NewCounterVec(
			config.counterOpts.apply(prometheus.CounterOpts{
				Name: "grpc_server_msg_received_total",
				Help: "Total number of RPC stream messages received on the server.",
			}), addIPLables([]string{"grpc_type", "grpc_service", "grpc_method"})),
		serverStreamMsgSent: prometheus.NewCounterVec(
			config.counterOpts.apply(prometheus.CounterOpts{
				Name: "grpc_server_msg_sent_total",
				Help: "Total number of gRPC stream messages sent by the server.",
			}), addIPLables([]string{"grpc_type", "grpc_service", "grpc_method"})),
	}
	if config.serverHandledHistogramEnabled {
		sm.serverHandledHistogram = newServerHandlingTimeHistogram(
			config.ipLabelsEnabled, config.serverHandledHistogramOptions)
	}

	return sm
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *ServerMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.serverStartedCounter.Describe(ch)
	m.serverHandledCounter.Describe(ch)
	m.serverStreamMsgReceived.Describe(ch)
	m.serverStreamMsgSent.Describe(ch)
	if m.serverHandledHistogram != nil {
		m.serverHandledHistogram.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *ServerMetrics) Collect(ch chan<- prometheus.Metric) {
	m.serverStartedCounter.Collect(ch)
	m.serverHandledCounter.Collect(ch)
	m.serverStreamMsgReceived.Collect(ch)
	m.serverStreamMsgSent.Collect(ch)
	if m.serverHandledHistogram != nil {
		m.serverHandledHistogram.Collect(ch)
	}
}

// InitializeMetrics initializes all metrics, with their appropriate null
// value, for all gRPC methods registered on a gRPC server. This is useful, to
// ensure that all metrics exist when collecting and querying.
// NOTE: This might add significant cardinality and might not be needed in future version of Prometheus (created timestamp).
func (m *ServerMetrics) InitializeMetrics(server reflection.ServiceInfoProvider) {
	serviceInfo := server.GetServiceInfo()
	for serviceName, info := range serviceInfo {
		for _, mInfo := range info.Methods {
			m.preRegisterMethod(serviceName, &mInfo)
		}
	}
}

// preRegisterMethod is invoked on Register of a Server, allowing all gRPC services labels to be pre-populated.
func (m *ServerMetrics) preRegisterMethod(serviceName string, mInfo *grpc.MethodInfo) {
	methodName := mInfo.Name
	methodType := string(typeFromMethodInfo(mInfo))

	lvals := []string{methodType, serviceName, methodName}
	if m.ipLabelsEnabled {
		// Because netip.Addr.String() returns "invalid IP" for zero Addr,
		// we use this value with grpc_server and grpc_client.
		lvals = append(lvals, "invalid IP", "invalid IP")
	}
	// These are just references (no increments), as just referencing will create the labels but not set values.
	_, _ = m.serverStartedCounter.GetMetricWithLabelValues(lvals...)
	_, _ = m.serverStreamMsgReceived.GetMetricWithLabelValues(lvals...)
	_, _ = m.serverStreamMsgSent.GetMetricWithLabelValues(lvals...)
	if m.serverHandledHistogram != nil {
		_, _ = m.serverHandledHistogram.GetMetricWithLabelValues(lvals...)
	}

	for _, code := range interceptors.AllCodes {
		lvals = []string{methodType, serviceName, methodName, code.String()}
		if m.ipLabelsEnabled {
			lvals = append(lvals, "invalid IP", "invalid IP")
		}
		_, _ = m.serverHandledCounter.GetMetricWithLabelValues(lvals...)
	}
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *ServerMetrics) UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&reportable{
		opts:          opts,
		serverMetrics: m,
	})
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *ServerMetrics) StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&reportable{
		opts:          opts,
		serverMetrics: m,
	})
}
