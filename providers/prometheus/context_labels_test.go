// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package prometheus

import (
	"context"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestServerContextLabels(t *testing.T) {
	// Create server metrics with context labels
	serverMetrics := NewServerMetrics(
		WithContextLabels("user_id", "tenant_id"),
	)

	// Create a custom registry to isolate this test
	reg := prometheus.NewRegistry()
	reg.MustRegister(serverMetrics)

	// Create a mock context with label values
	ctx := context.Background()

	// Create the labels extraction function
	labelsFromCtx := func(ctx context.Context) prometheus.Labels {
		return prometheus.Labels{
			"user_id":   "user123",
			"tenant_id": "tenant456",
		}
	}

	// Create a reporter with the labels function
	rep := &reportable{
		serverMetrics: serverMetrics,
		opts:          []Option{WithLabelsFromContext(labelsFromCtx)},
	}

	// Simulate a server call
	meta := interceptors.CallMeta{
		Typ:     interceptors.Unary,
		Service: "testpb.PingService",
		Method:  "Ping",
	}

	reporter, _ := rep.ServerReporter(ctx, meta)

	// Simulate call completion
	reporter.PostCall(nil, time.Millisecond*100)

	// Collect metrics
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	// Find the handled counter metric
	var handledCounter *dto.MetricFamily
	for _, mf := range metricFamilies {
		if *mf.Name == "grpc_server_handled_total" {
			handledCounter = mf
			break
		}
	}

	require.NotNil(t, handledCounter, "Should find grpc_server_handled_total metric")
	require.Len(t, handledCounter.Metric, 1, "Should have one metric sample")

	// Verify the metric has all expected labels
	metric := handledCounter.Metric[0]
	labelMap := make(map[string]string)
	for _, label := range metric.Label {
		labelMap[*label.Name] = *label.Value
	}

	// Check standard labels
	assert.Equal(t, "unary", labelMap["grpc_type"])
	assert.Equal(t, "testpb.PingService", labelMap["grpc_service"])
	assert.Equal(t, "Ping", labelMap["grpc_method"])
	assert.Equal(t, codes.OK.String(), labelMap["grpc_code"])

	// Check context labels
	assert.Equal(t, "user123", labelMap["user_id"])
	assert.Equal(t, "tenant456", labelMap["tenant_id"])

	// Verify metric value
	assert.InDelta(t, float64(1), *metric.Counter.Value, 0.0001, "Metric value should be 1")
}

func TestClientContextLabels(t *testing.T) {
	// Create client metrics with context labels
	clientMetrics := NewClientMetrics(
		WithClientContextLabels("user_id", "tenant_id"),
	)

	// Create a custom registry to isolate this test
	reg := prometheus.NewRegistry()
	reg.MustRegister(clientMetrics)

	// Create a mock context with label values
	ctx := context.Background()

	// Create the labels extraction function
	labelsFromCtx := func(ctx context.Context) prometheus.Labels {
		return prometheus.Labels{
			"user_id":   "user123",
			"tenant_id": "tenant456",
		}
	}

	// Create a reporter with the labels function
	rep := &reportable{
		clientMetrics: clientMetrics,
		opts:          []Option{WithLabelsFromContext(labelsFromCtx)},
	}

	// Simulate a server call
	meta := interceptors.CallMeta{
		Typ:     interceptors.Unary,
		Service: "testpb.PingService",
		Method:  "Ping",
	}

	reporter, _ := rep.ClientReporter(ctx, meta)

	// Simulate call completion
	reporter.PostCall(nil, time.Millisecond*100)

	// Collect metrics
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	// Find the handled counter metric
	var handledCounter *dto.MetricFamily
	for _, mf := range metricFamilies {
		if *mf.Name == "grpc_client_handled_total" {
			handledCounter = mf
			break
		}
	}

	require.NotNil(t, handledCounter, "Should find grpc_server_handled_total metric")
	require.Len(t, handledCounter.Metric, 1, "Should have one metric sample")

	// Verify the metric has all expected labels
	metric := handledCounter.Metric[0]
	labelMap := make(map[string]string)
	for _, label := range metric.Label {
		labelMap[*label.Name] = *label.Value
	}

	// Check standard labels
	assert.Equal(t, "unary", labelMap["grpc_type"])
	assert.Equal(t, "testpb.PingService", labelMap["grpc_service"])
	assert.Equal(t, "Ping", labelMap["grpc_method"])
	assert.Equal(t, codes.OK.String(), labelMap["grpc_code"])

	// Check context labels
	assert.Equal(t, "user123", labelMap["user_id"])
	assert.Equal(t, "tenant456", labelMap["tenant_id"])

	// Verify metric value
	assert.InDelta(t, float64(1), *metric.Counter.Value, 0.0001, "Metric value should be 1")
}

func TestServerContextLabelsWithMissingValues(t *testing.T) {
	// Create server metrics with context labels
	serverMetrics := NewServerMetrics(
		WithContextLabels("user_id", "missing_label"),
	)

	// Create a custom registry to isolate this test
	reg := prometheus.NewRegistry()
	reg.MustRegister(serverMetrics)

	// Create a mock context with only partial label values
	ctx := context.Background()

	// Create the labels extraction function that only returns one label
	labelsFromCtx := func(ctx context.Context) prometheus.Labels {
		return prometheus.Labels{
			"user_id": "user123",
			// missing_label is not provided
		}
	}

	// Create a reporter with the labels function
	rep := &reportable{
		serverMetrics: serverMetrics,
		opts:          []Option{WithLabelsFromContext(labelsFromCtx)},
	}

	// Simulate a server call
	meta := interceptors.CallMeta{
		Typ:     interceptors.Unary,
		Service: "testpb.PingService",
		Method:  "Ping",
	}

	reporter, _ := rep.ServerReporter(ctx, meta)

	// Simulate call completion
	reporter.PostCall(nil, time.Millisecond*100)

	// Collect metrics
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	// Find the handled counter metric
	var handledCounter *dto.MetricFamily
	for _, mf := range metricFamilies {
		if *mf.Name == "grpc_server_handled_total" {
			handledCounter = mf
			break
		}
	}

	require.NotNil(t, handledCounter, "Should find grpc_server_handled_total metric")
	require.Len(t, handledCounter.Metric, 1, "Should have one metric sample")

	// Verify the metric has all expected labels
	metric := handledCounter.Metric[0]
	labelMap := make(map[string]string)
	for _, label := range metric.Label {
		labelMap[*label.Name] = *label.Value
	}

	// Check standard labels
	assert.Equal(t, "unary", labelMap["grpc_type"])
	assert.Equal(t, "testpb.PingService", labelMap["grpc_service"])
	assert.Equal(t, "Ping", labelMap["grpc_method"])
	assert.Equal(t, codes.OK.String(), labelMap["grpc_code"])

	// Check context labels - user_id should be present, missing_label should be empty
	assert.Equal(t, "user123", labelMap["user_id"])
	assert.Empty(t, labelMap["missing_label"])

	// Verify metric value
	assert.InDelta(t, float64(1), *metric.Counter.Value, 0.0001, "Metric value should be 1")
}

func TestClientContextLabelsWithMissingValues(t *testing.T) {
	// Create client metrics with context labels
	clientMetrics := NewClientMetrics(
		WithClientContextLabels("user_id", "missing_label"),
	)

	// Create a custom registry to isolate this test
	reg := prometheus.NewRegistry()
	reg.MustRegister(clientMetrics)

	// Create a mock context with only partial label values
	ctx := context.Background()

	// Create the labels extraction function that only returns one label
	labelsFromCtx := func(ctx context.Context) prometheus.Labels {
		return prometheus.Labels{
			"user_id": "user123",
			// missing_label is not provided
		}
	}

	// Create a reporter with the labels function
	rep := &reportable{
		clientMetrics: clientMetrics,
		opts:          []Option{WithLabelsFromContext(labelsFromCtx)},
	}

	// Simulate a server call
	meta := interceptors.CallMeta{
		Typ:     interceptors.Unary,
		Service: "testpb.PingService",
		Method:  "Ping",
	}

	reporter, _ := rep.ClientReporter(ctx, meta)

	// Simulate call completion
	reporter.PostCall(nil, time.Millisecond*100)

	// Collect metrics
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	// Find the handled counter metric
	var handledCounter *dto.MetricFamily
	for _, mf := range metricFamilies {
		if *mf.Name == "grpc_client_handled_total" {
			handledCounter = mf
			break
		}
	}

	require.NotNil(t, handledCounter, "Should find grpc_server_handled_total metric")
	require.Len(t, handledCounter.Metric, 1, "Should have one metric sample")

	// Verify the metric has all expected labels
	metric := handledCounter.Metric[0]
	labelMap := make(map[string]string)
	for _, label := range metric.Label {
		labelMap[*label.Name] = *label.Value
	}

	// Check standard labels
	assert.Equal(t, "unary", labelMap["grpc_type"])
	assert.Equal(t, "testpb.PingService", labelMap["grpc_service"])
	assert.Equal(t, "Ping", labelMap["grpc_method"])
	assert.Equal(t, codes.OK.String(), labelMap["grpc_code"])

	// Check context labels - user_id should be present, missing_label should be empty
	assert.Equal(t, "user123", labelMap["user_id"])
	assert.Empty(t, labelMap["missing_label"])

	// Verify metric value
	assert.InDelta(t, float64(1), *metric.Counter.Value, 0.0001, "Metric value should be 1")
}

func TestServerContextLabelsWithHistogram(t *testing.T) {
	// Create server metrics with context labels and histogram enabled
	serverMetrics := NewServerMetrics(
		WithContextLabels("user_id"),
		WithServerHandlingTimeHistogram(),
	)

	// Create a custom registry to isolate this test
	reg := prometheus.NewRegistry()
	reg.MustRegister(serverMetrics)

	// Create a mock context with label values
	ctx := context.Background()

	// Create the labels extraction function
	labelsFromCtx := func(ctx context.Context) prometheus.Labels {
		return prometheus.Labels{
			"user_id": "user123",
		}
	}

	// Create a reporter with the labels function
	rep := &reportable{
		serverMetrics: serverMetrics,
		opts:          []Option{WithLabelsFromContext(labelsFromCtx)},
	}

	// Simulate a server call
	meta := interceptors.CallMeta{
		Typ:     interceptors.Unary,
		Service: "testpb.PingService",
		Method:  "Ping",
	}

	reporter, _ := rep.ServerReporter(ctx, meta)

	// Simulate call completion
	reporter.PostCall(nil, time.Millisecond*100)

	// Collect metrics
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	// Find the histogram metric
	var histogram *dto.MetricFamily
	for _, mf := range metricFamilies {
		if *mf.Name == "grpc_server_handling_seconds" {
			histogram = mf
			break
		}
	}

	require.NotNil(t, histogram, "Should find grpc_server_handling_seconds metric")
	require.Len(t, histogram.Metric, 1, "Should have one metric sample")

	// Verify the histogram has all expected labels
	metric := histogram.Metric[0]
	labelMap := make(map[string]string)
	for _, label := range metric.Label {
		labelMap[*label.Name] = *label.Value
	}

	// Check standard labels
	assert.Equal(t, "unary", labelMap["grpc_type"])
	assert.Equal(t, "testpb.PingService", labelMap["grpc_service"])
	assert.Equal(t, "Ping", labelMap["grpc_method"])

	// Check context labels
	assert.Equal(t, "user123", labelMap["user_id"])

	// Verify histogram has recorded a sample
	assert.Equal(t, uint64(1), *metric.Histogram.SampleCount)
}

func TestClientContextLabelsWithHistogram(t *testing.T) {
	// Create client metrics with context labels and histogram enabled
	clientMetrics := NewClientMetrics(
		WithClientContextLabels("user_id"),
		WithClientHandlingTimeHistogram(),
	)

	// Create a custom registry to isolate this test
	reg := prometheus.NewRegistry()
	reg.MustRegister(clientMetrics)

	// Create a mock context with label values
	ctx := context.Background()

	// Create the labels extraction function
	labelsFromCtx := func(ctx context.Context) prometheus.Labels {
		return prometheus.Labels{
			"user_id": "user123",
		}
	}

	// Create a reporter with the labels function
	rep := &reportable{
		clientMetrics: clientMetrics,
		opts:          []Option{WithLabelsFromContext(labelsFromCtx)},
	}

	// Simulate a server call
	meta := interceptors.CallMeta{
		Typ:     interceptors.Unary,
		Service: "testpb.PingService",
		Method:  "Ping",
	}

	reporter, _ := rep.ClientReporter(ctx, meta)

	// Simulate call completion
	reporter.PostCall(nil, time.Millisecond*100)

	// Collect metrics
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	// Find the histogram metric
	var histogram *dto.MetricFamily
	for _, mf := range metricFamilies {
		if *mf.Name == "grpc_client_handling_seconds" {
			histogram = mf
			break
		}
	}

	require.NotNil(t, histogram, "Should find grpc_server_handling_seconds metric")
	require.Len(t, histogram.Metric, 1, "Should have one metric sample")

	// Verify the histogram has all expected labels
	metric := histogram.Metric[0]
	labelMap := make(map[string]string)
	for _, label := range metric.Label {
		labelMap[*label.Name] = *label.Value
	}

	// Check standard labels
	assert.Equal(t, "unary", labelMap["grpc_type"])
	assert.Equal(t, "testpb.PingService", labelMap["grpc_service"])
	assert.Equal(t, "Ping", labelMap["grpc_method"])

	// Check context labels
	assert.Equal(t, "user123", labelMap["user_id"])

	// Verify histogram has recorded a sample
	assert.Equal(t, uint64(1), *metric.Histogram.SampleCount)
}
