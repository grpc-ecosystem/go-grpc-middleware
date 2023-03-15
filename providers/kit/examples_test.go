// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package kit_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/kit"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

var (
	customFunc             logging.CodeToLevel
	customDurationToFields logging.DurationToFields
)

func Example_initializationWithCustomLevels() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := log.NewNopLogger()
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []logging.Option{
		logging.WithLevels(customFunc),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(kit.InterceptorLogger(logger), opts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(kit.InterceptorLogger(logger), opts...),
		),
	)
}

func Example_initializationWithDurationFieldOverride() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := log.NewNopLogger()
	// Shared options for the logger, with a custom duration to log field function.
	opts := []logging.Option{
		logging.WithDurationField(customDurationToFields),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(kit.InterceptorLogger(logger), opts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(kit.InterceptorLogger(logger), opts...),
		),
	)
}

func Example_initializationWithCodeGenRequestFieldExtractor() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := log.NewNopLogger()
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(kit.InterceptorLogger(logger)),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(kit.InterceptorLogger(logger)),
		),
	)
}

func ExampleWithDecider() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := log.NewNopLogger()
	// Shared options for the logger, with a custom decider that log everything except successful calls from "/blah.foo.healthcheck/Check" method.
	opts := []logging.Option{
		logging.WithDecider(func(methodFullName string, err error) logging.Decision {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if methodFullName == "/blah.foo.healthcheck/Check" {
				return logging.NoLogCall
			}
			// by default you will log all calls
			return logging.LogStartAndFinishCall
		}),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(kit.InterceptorLogger(logger), opts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(kit.InterceptorLogger(logger), opts...),
		),
	}
}

func ExampleServerPayloadLoggingDecider() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := log.NewNopLogger()
	// Expect payload from  "/blah.foo.healthcheck/Check" call to be logged.
	payloadDecider := func(ctx context.Context, fullMethodName string, servingObject interface{}) logging.PayloadDecision {
		// return fullMethodName == "/blah.foo.healthcheck/Check"
		// TODO Fix
		return 0
	}

	// Create a server, make sure we put the tags context before everything else.
	_ = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(kit.InterceptorLogger(logger)),
			logging.PayloadUnaryServerInterceptor(kit.InterceptorLogger(logger), payloadDecider, time.RFC3339Nano),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(kit.InterceptorLogger(logger)),
			logging.PayloadStreamServerInterceptor(kit.InterceptorLogger(logger), payloadDecider, time.RFC3339Nano),
		),
	}
}

func TestExamplesBuildable(t *testing.T) {
	Example_initializationWithCustomLevels()
	Example_initializationWithDurationFieldOverride()
	Example_initializationWithCodeGenRequestFieldExtractor()
	ExampleWithDecider()
	ExampleServerPayloadLoggingDecider()
}
