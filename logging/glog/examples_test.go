// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog_test

import (
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	logger     grpclog.LoggerV2
	customFunc grpc_glog.CodeToLevel
)

// Initialization shows a relatively complex initialization sequence.
func Example_Initialization() {
	// glog entry is used, allowing pre-definition of certain fields by the user.
	glogEntry := ctx_glog.NewEntry(logger)
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_glog.Option{
		grpc_glog.WithLevels(customFunc),
	}
	// Make sure that log statements internal to gRPC library are logged using the glog as well.
	grpc_glog.ReplaceGrpcLogger()
	// Create a server, make sure we put the grpc_ctxtags context before everything else.
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_glog.UnaryServerInterceptor(glogEntry, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_glog.StreamServerInterceptor(glogEntry, opts...),
		),
	)
}

func Example_InitializationWithDurationFieldOverride() {
	// glog entry is used, allowing pre-definition of certain fields by the user.
	glogEntry := ctx_glog.NewEntry(logger)
	// Shared options for the logger, with a custom duration to log field function.
	opts := []grpc_glog.Option{
		grpc_glog.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_glog.UnaryServerInterceptor(glogEntry, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_glog.StreamServerInterceptor(glogEntry, opts...),
		),
	)
}

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func Example_HandlerUsageUnaryPing() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
		// Extract a single request-scoped glog.Logger and log messages.
		l := ctx_glog.Extract(ctx)
		l.Info("some ping")
		l.Info("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}

func ExampleWithDecider() {
	opts := []grpc_glog.Option{
		grpc_glog.WithDecider(func(methodFullName string, err error) bool {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if err == nil && methodFullName == "blah.foo.healthcheck" {
				return false
			}

			// by default you will log all calls
			return true
		}),
	}

	_ = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_glog.StreamServerInterceptor(ctx_glog.NewEntry(nullLogger), opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_glog.UnaryServerInterceptor(ctx_glog.NewEntry(nullLogger), opts...)),
	}
}
