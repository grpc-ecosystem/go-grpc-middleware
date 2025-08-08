// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcMetadata "google.golang.org/grpc/metadata"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

const (
	component      = "grpc-example"
	httpAddr       = ":8082"
	targetGRPCAddr = "localhost:8080"
)

// interceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func main() {
	// Setup logging.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	rpcLogger := logger.With("service", "gRPC/client", "component", component)
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}

	// Setup metrics.
	reg := prometheus.NewRegistry()
	clMetrics := grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg.MustRegister(clMetrics)
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	// Set up OTLP tracing (stdout for debug).
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		logger.Error("failed to init exporter", "err", err)
		os.Exit(1)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	defer func() { _ = exporter.Shutdown(context.Background()) }()

	cc, err := grpc.NewClient(
		targetGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(500*time.Millisecond),
			clMetrics.UnaryClientInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.UnaryClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID))),
		grpc.WithChainStreamInterceptor(
			clMetrics.StreamClientInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.StreamClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID))),
	)
	if err != nil {
		logger.Error("failed to init gRPC client", "err", err)
		os.Exit(1)
	}

	cl := testpb.NewTestServiceClient(cc)

	g := &run.Group{}
	ctx, cancel := context.WithCancel(context.Background())
	g.Add(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(1 * time.Second):
			}

			md := grpcMetadata.Pairs("authorization", "bearer yolo")
			if _, err := cl.Ping(metadata.MD(md).ToOutgoing(ctx), &testpb.PingRequest{Value: "example"}); err != nil {
				return err
			}
		}
	}, func(err error) {
		cancel()
	})

	httpSrv := &http.Server{Addr: httpAddr}
	g.Add(func() error {
		m := http.NewServeMux()
		// Create HTTP handler for Prometheus metrics.
		m.Handle("/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				// Opt into OpenMetrics e.g. to support exemplars.
				EnableOpenMetrics: true,
			},
		))
		httpSrv.Handler = m
		logger.Info("starting HTTP server", "addr", httpSrv.Addr)
		return httpSrv.ListenAndServe()
	}, func(error) {
		if err := httpSrv.Close(); err != nil {
			logger.Error("failed to stop web server", "err", err)
		}
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		logger.Error("program interrupted", "err", err)
		os.Exit(1)
	}
}
