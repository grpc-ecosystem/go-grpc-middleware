// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logrus_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-grpc-middleware/logging/logrus"

	"github.com/mwitkow/go-grpc-middleware/logging"
	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Initialization shows a relatively complex initialization sequence.
func Example_initialization(logrusLogger *logrus.Logger, customFunc grpc_logrus.CodeToLevel) *grpc.Server {
	// Logrus entry is used, allowing pre-definition of certain fields by the user.
	logrusEntry := logrus.NewEntry(logrusLogger)
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(customFunc),
		grpc_logrus.WithFieldExtractor(grpc_logging.CodeGenRequestLogFieldExtractor), // default, don't have to set
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)
	// Create a server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_logrus.UnaryServerInterceptor(logrusEntry, opts...)),
		grpc.StreamInterceptor(grpc_logrus.StreamServerInterceptor(logrusEntry, opts...)),
	)
	return server
}

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func Example_handlerUsageUnaryPing() interface{} {
	x := func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields to all log statements, including the final one made by the interceptor.
		grpc_logrus.AddFields(ctx, logrus.Fields{"custom_string": "something", "custom_int": 1337})
		// Extract a request-scoped zap.Logger and log a message.
		grpc_logrus.Extract(ctx).Info("some ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
	return x
}
