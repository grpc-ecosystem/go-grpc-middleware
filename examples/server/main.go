// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"syscall"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

const (
	component = "grpc-example"
	grpcAddr  = ":8080"
	httpAddr  = ":8081"
)

// interceptorLogger adapts go-kit logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func interceptorLogger(l log.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		largs := append([]any{"msg", msg}, fields...)
		switch lvl {
		case logging.LevelDebug:
			_ = level.Debug(l).Log(largs...)
		case logging.LevelInfo:
			_ = level.Info(l).Log(largs...)
		case logging.LevelWarn:
			_ = level.Warn(l).Log(largs...)
		case logging.LevelError:
			_ = level.Error(l).Log(largs...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func main() {
	// Setup logging.
	logger := log.NewLogfmtLogger(os.Stderr)
	rpcLogger := log.With(logger, "service", "gRPC/server", "component", component)
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}

	// Setup metrics.
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	// Set up OTLP tracing (stdout for debug).
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	defer func() { _ = exporter.Shutdown(context.Background()) }()

	// Setup custom auth.
	authFn := func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		// TODO: This is example only, perform proper Oauth/OIDC verification!
		if token != "yolo" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
		}
		// NOTE: You can also pass the token in the context for further interceptors or gRPC service code.
		return ctx, nil
	}

	// Setup auth matcher.
	allButHealthZ := func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		return healthpb.Health_ServiceDesc.ServiceName != callMeta.Service
	}

	// Setup metric for panic recoveries.
	panicsTotal := promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})
	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		level.Error(rpcLogger).Log("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
			otelgrpc.UnaryServerInterceptor(),
			srvMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.UnaryServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(authFn), selector.MatchFunc(allButHealthZ)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			otelgrpc.StreamServerInterceptor(),
			srvMetrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.StreamServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			selector.StreamServerInterceptor(auth.StreamServerInterceptor(authFn), selector.MatchFunc(allButHealthZ)),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	)
	t := &testpb.TestPingService{}
	testpb.RegisterTestServiceServer(grpcSrv, t)
	srvMetrics.InitializeMetrics(grpcSrv)

	g := &run.Group{}
	g.Add(func() error {
		l, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			return err
		}
		level.Info(logger).Log("msg", "starting gRPC server", "addr", l.Addr().String())
		return grpcSrv.Serve(l)
	}, func(err error) {
		grpcSrv.GracefulStop()
		grpcSrv.Stop()
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
		level.Info(logger).Log("msg", "starting HTTP server", "addr", httpSrv.Addr)
		return httpSrv.ListenAndServe()
	}, func(error) {
		if err := httpSrv.Close(); err != nil {
			level.Error(logger).Log("msg", "failed to stop web server", "err", err)
		}
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
