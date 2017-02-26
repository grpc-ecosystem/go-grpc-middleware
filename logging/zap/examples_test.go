// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zap_test

import (
	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"

	"context"

	"github.com/mwitkow/go-grpc-middleware/logging"
	"github.com/mwitkow/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Initialization shows a relatively complex initialization sequence.
func Example_initialization(zapLogger *zap.Logger, customFunc grpc_zap.CodeToLevel) *grpc.Server {
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_zap.Option{
		grpc_zap.WithLevels(customFunc),
		grpc_zap.WithFieldExtractor(grpc_logging.CodeGenRequestLogFieldExtractor), // default, don't have to set
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLogger(zapLogger)
	// Create a server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_zap.UnaryServerInterceptor(zapLogger, opts...)),
		grpc.StreamInterceptor(grpc_zap.StreamServerInterceptor(zapLogger, opts...)),
	)
	return server
}

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func Example_handlerUsageUnaryPing() interface{} {
	x := func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields to all log statements, including the final one made by the interceptor.
		grpc_zap.AddFields(ctx, zap.String("custom_string", "something"), zap.Int("custom_int", 1337))
		// Extract a request-scoped zap.Logger and log a message.
		grpc_zap.Extract(ctx).Info("some ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
	return x
}
