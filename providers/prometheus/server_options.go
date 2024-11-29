// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package prometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func serverMetricAddIPLabel(orig []string) []string {
	return append(orig, "grpc_server_ip", "grpc_client_ip")
}

type exemplarFromCtxFn func(ctx context.Context) prometheus.Labels

type serverMetricsConfig struct {
	// ipLabelsEnabled control whether to add grpc_server and grpc_client labels to metrics.
	ipLabelsEnabled bool

	counterOpts counterOptions

	serverHandledHistogramEnabled bool
	serverHandledHistogramOptions []HistogramOption
}

// ServerMetricsOption configures how we set up the server metrics.
type ServerMetricsOption func(*serverMetricsConfig)

func (c *serverMetricsConfig) apply(opts []ServerMetricsOption) {
	for _, o := range opts {
		o(c)
	}
}

// WithServerCounterOptions sets counter options.
func WithServerCounterOptions(opts ...CounterOption) ServerMetricsOption {
	return func(o *serverMetricsConfig) {
		o.counterOpts = opts
	}
}

func newServerHandlingTimeHistogram(ipLabelsEnabled bool, opts []HistogramOption) *prometheus.HistogramVec {
	labels := []string{"grpc_type", "grpc_service", "grpc_method"}
	if ipLabelsEnabled {
		labels = serverMetricAddIPLabel(labels)
	}
	return prometheus.NewHistogramVec(
		histogramOptions(opts).apply(prometheus.HistogramOpts{
			Name:    "grpc_server_handling_seconds",
			Help:    "Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.",
			Buckets: prometheus.DefBuckets,
		}),
		labels,
	)
}

// WithServerHandlingTimeHistogram turns on recording of handling time of RPCs.
// Histogram metrics can be very expensive for Prometheus to retain and query.
func WithServerHandlingTimeHistogram(opts ...HistogramOption) ServerMetricsOption {
	return func(o *serverMetricsConfig) {
		o.serverHandledHistogramEnabled = true
		o.serverHandledHistogramOptions = opts
	}
}

// WithServerIPLabelsEnabled enables adding grpc_server and grpc_client labels to metrics.
func WithServerIPLabelsEnabled() ServerMetricsOption {
	return func(o *serverMetricsConfig) {
		o.ipLabelsEnabled = true
	}
}
