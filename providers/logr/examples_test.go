// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logr_test

import (
	"context"
	"testing"

	grpclogr "github.com/grpc-ecosystem/go-grpc-middleware/providers/logr/v2"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

var (
	customFunc             logging.CodeToLevel
	customDurationToFields logging.DurationToFields
)

func Example_initializationWithCustomLevels() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := klog.NewKlogr()
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []logging.Option{
		logging.WithLevels(customFunc),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		middleware.WithUnaryServerChain(
			logging.UnaryServerInterceptor(grpclogr.InterceptorLogger(logger), opts...),
		),
		middleware.WithStreamServerChain(
			logging.StreamServerInterceptor(grpclogr.InterceptorLogger(logger), opts...),
		),
	)
}

func Example_initializationWithDurationFieldOverride() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := klog.NewKlogr()
	// Shared options for the logger, with a custom duration to log field function.
	opts := []logging.Option{
		logging.WithDurationField(customDurationToFields),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		middleware.WithUnaryServerChain(
			logging.UnaryServerInterceptor(grpclogr.InterceptorLogger(logger), opts...),
		),
		middleware.WithStreamServerChain(
			logging.StreamServerInterceptor(grpclogr.InterceptorLogger(logger), opts...),
		),
	)
}

func ExampleWithDecider() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := klog.NewKlogr()
	// Shared options for the logger, with a custom decider that log everything except successful calls from "/blah.foo.healthcheck/Check" method.
	opts := []logging.Option{
		logging.WithDecider(func(methodFullName string) logging.Decision {
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
		middleware.WithUnaryServerChain(
			logging.UnaryServerInterceptor(grpclogr.InterceptorLogger(logger), opts...),
		),
		middleware.WithStreamServerChain(
			logging.StreamServerInterceptor(grpclogr.InterceptorLogger(logger), opts...),
		),
	}
}

func ExampleServerPayloadLoggingDecider() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := klog.NewKlogr()
	// Expect payload from  "/blah.foo.healthcheck/Check" call to be logged.
	payloadDecider := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
		return fullMethodName == "/blah.foo.healthcheck/Check"
	}

	// Create a server, make sure we put the tags context before everything else.
	_ = []grpc.ServerOption{
		middleware.WithUnaryServerChain(
			logging.UnaryServerInterceptor(grpclogr.InterceptorLogger(logger)),
			logging.PayloadUnaryServerInterceptor(grpclogr.InterceptorLogger(logger), payloadDecider),
		),
		middleware.WithStreamServerChain(
			logging.StreamServerInterceptor(grpclogr.InterceptorLogger(logger)),
			logging.PayloadStreamServerInterceptor(grpclogr.InterceptorLogger(logger), payloadDecider),
		),
	}
}

func TestExamplesBuildable(t *testing.T) {
	Example_initializationWithCustomLevels()
	Example_initializationWithDurationFieldOverride()
	ExampleWithDecider()
	ExampleServerPayloadLoggingDecider()
}
