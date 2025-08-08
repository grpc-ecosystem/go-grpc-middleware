// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package slog_test

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
)

func InterceptorLogger() logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		fmt.Printf("This will not print anything for health checks or reflection, but will print this message for other requests")
	})
}

func SkipHealthAndReflectionRequests(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() != "/grpc.health.v1.Health/Check" &&
		c.FullMethod() != "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"
}

func ExampleInterceptorLogger() {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	// Create a server and wrap the logging interceptor with the selector interceptor to skip health checks and reflection requests.
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			// Wrap the logging interceptor with the selector interceptor to skip health checks and reflection requests.
			selector.UnaryServerInterceptor(
				logging.UnaryServerInterceptor(InterceptorLogger(), opts...),
				selector.MatchFunc(SkipHealthAndReflectionRequests),
			),
		),
		grpc.ChainStreamInterceptor(
			selector.StreamServerInterceptor(
				logging.StreamServerInterceptor(InterceptorLogger(), opts...),
				selector.MatchFunc(SkipHealthAndReflectionRequests),
			),
		),
	)

	// Similarly you can create client and wrap the logging interceptor with the selector interceptor to skip health checks and reflection requests.
	_, _ = grpc.Dial(
		"some-target",
		grpc.WithChainUnaryInterceptor(
			selector.UnaryClientInterceptor(
				logging.UnaryClientInterceptor(InterceptorLogger(), opts...),
				selector.MatchFunc(SkipHealthAndReflectionRequests),
			),
		),
		grpc.WithChainStreamInterceptor(
			selector.StreamClientInterceptor(
				logging.StreamClientInterceptor(InterceptorLogger(), opts...),
				selector.MatchFunc(SkipHealthAndReflectionRequests),
			),
		),
	)
	// Output:
}
