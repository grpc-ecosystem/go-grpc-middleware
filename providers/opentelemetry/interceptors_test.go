// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentelemetry_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	grpcopentelemetry "github.com/grpc-ecosystem/go-grpc-middleware/providers/opentelemetry/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

var (
	fakeInboundTraceId = "4bf92f3577b34da6a3ce929d0e0e4736"
	fakeInboundSpanId  = "00f067aa0ba902b7"
)

type tracingAssertService struct {
	testpb.TestServiceServer
	T *testing.T
}

func (s *tracingAssertService) Ping(ctx context.Context, ping *testpb.PingRequest) (*testpb.PingResponse, error) {
	assert.NotNil(s.T, trace.SpanFromContext(ctx), "handlers must have the spancontext in their context, otherwise propagation will fail")

	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *tracingAssertService) PingError(ctx context.Context, ping *testpb.PingErrorRequest) (*testpb.PingErrorResponse, error) {
	assert.NotNil(s.T, trace.SpanFromContext(ctx), "handlers must have the spancontext in their context, otherwise propagation will fail")
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *tracingAssertService) PingList(ping *testpb.PingListRequest, stream testpb.TestService_PingListServer) error {
	assert.NotNil(s.T, trace.SpanFromContext(stream.Context()), "handlers must have the spancontext in their context, otherwise propagation will fail")

	return s.TestServiceServer.PingList(ping, stream)
}

func (s *tracingAssertService) PingEmpty(ctx context.Context, empty *testpb.PingEmptyRequest) (*testpb.PingEmptyResponse, error) {
	assert.NotNil(s.T, trace.SpanFromContext(ctx), "handlers must have the spancontext in their context, otherwise propagation will fail")

	return s.TestServiceServer.PingEmpty(ctx, empty)
}

func TestProvider(t *testing.T) {
	var srd SpanRecorderDelegate
	suite.Run(t, &OpenTelemetrySuite{
		srd: &srd,
		InterceptorTestSuite: makeInterceptorTestSuite(t, []grpcopentelemetry.Option{
			grpcopentelemetry.WithTracerProvider(sdktrace.NewTracerProvider(
				sdktrace.WithSpanProcessor(&srd),
			)),
			grpcopentelemetry.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})),
		}),
	})
}

func makeInterceptorTestSuite(t *testing.T, opts []grpcopentelemetry.Option) *testpb.InterceptorTestSuite {
	return &testpb.InterceptorTestSuite{
		TestService: &tracingAssertService{TestServiceServer: &testpb.TestPingService{T: t}, T: t},
		ClientOpts: []grpc.DialOption{
			grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor(grpcopentelemetry.InterceptorTracer(opts...))),
			grpc.WithStreamInterceptor(tracing.StreamClientInterceptor(grpcopentelemetry.InterceptorTracer(opts...))),
		},
		ServerOpts: []grpc.ServerOption{
			grpc.ChainUnaryInterceptor(
				tracing.UnaryServerInterceptor(grpcopentelemetry.InterceptorTracer(opts...)),
			),
			grpc.ChainStreamInterceptor(
				tracing.StreamServerInterceptor(grpcopentelemetry.InterceptorTracer(opts...)),
			),
		},
	}
}

type SpanRecorderDelegate struct {
	sr *tracetest.SpanRecorder
}

func (s *SpanRecorderDelegate) OnStart(parent context.Context, span sdktrace.ReadWriteSpan) {
	s.sr.OnStart(parent, span)
}

func (s *SpanRecorderDelegate) OnEnd(span sdktrace.ReadOnlySpan) {
	s.sr.OnEnd(span)
}

func (s *SpanRecorderDelegate) Shutdown(ctx context.Context) error {
	return s.sr.Shutdown(ctx)
}

func (s *SpanRecorderDelegate) ForceFlush(ctx context.Context) error {
	return s.sr.ForceFlush(ctx)
}

type OpenTelemetrySuite struct {
	*testpb.InterceptorTestSuite
	srd *SpanRecorderDelegate
}

func (s *OpenTelemetrySuite) SetupTest() {
	var sr tracetest.SpanRecorder

	s.srd.sr = &sr
}

func (s *OpenTelemetrySuite) createContextFromFakeHttpRequestParent(ctx context.Context, sampled bool) context.Context {
	var flags trace.TraceFlags
	if sampled {
		flags = trace.FlagsSampled
	}

	return trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    mustTraceIDFromHex(fakeInboundTraceId),
		SpanID:     mustSpanIDFromHex(fakeInboundSpanId),
		TraceFlags: flags,
	}))
}

func (s *OpenTelemetrySuite) assertTracesCreated(methodName string) {
	spans := s.srd.sr.Ended()
	for _, span := range spans {
		s.T().Logf("span: %v, attributes: %v", span, span.Attributes())
	}

	require.Len(s.T(), spans, 2)
	for _, span := range spans {
		assert.Equal(s.T(), fakeInboundTraceId, span.SpanContext().TraceID().String())
		assert.Len(s.T(), span.Attributes(), 3)

		attributes := []attribute.KeyValue{
			semconv.RPCSystemGRPC,
			semconv.RPCServiceKey.String(testpb.TestServiceFullName),
			semconv.RPCMethodKey.String(methodName),
		}
		for _, v := range span.Attributes() {
			assert.Contains(s.T(), attributes, v)
		}
	}
}

func (s *OpenTelemetrySuite) TestPing_PropagatesTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), true)
	goodPing := testpb.PingRequest{Value: "something", SleepTimeMs: 9999}
	_, err := s.Client.Ping(ctx, &goodPing)
	require.NoError(s.T(), err, "there must be not be an on a successful call")
	s.assertTracesCreated("Ping")
}

func (s *OpenTelemetrySuite) TestPingList_PropagatesTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), true)
	goodPing := testpb.PingListRequest{Value: "something", SleepTimeMs: 9999}
	stream, err := s.Client.PingList(ctx, &goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}
	s.assertTracesCreated("PingList")
}

func (s *OpenTelemetrySuite) TestPingError_PropagatesTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), true)
	erroringPing := testpb.PingErrorRequest{Value: "something", ErrorCodeReturned: uint32(codes.OutOfRange)}
	_, err := s.Client.PingError(ctx, &erroringPing)
	require.Error(s.T(), err, "there must be an error returned here")
}

func (s *OpenTelemetrySuite) TestPingEmpty_NotSampleTraces() {
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx(), false)
	_, err := s.Client.PingEmpty(ctx, &testpb.PingEmptyRequest{})
	require.NoError(s.T(), err, "there must be not be an on a successful call")
}

func mustTraceIDFromHex(s string) (t trace.TraceID) {
	var err error
	t, err = trace.TraceIDFromHex(s)
	if err != nil {
		panic(err)
	}
	return
}

func mustSpanIDFromHex(s string) (t trace.SpanID) {
	var err error
	t, err = trace.SpanIDFromHex(s)
	if err != nil {
		panic(err)
	}
	return
}
