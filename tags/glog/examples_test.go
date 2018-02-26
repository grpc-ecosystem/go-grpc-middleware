// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package ctx_glog_test

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"golang.org/x/net/context"
)

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func Example_HandlerUsageUnaryPing() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
		// Extract a single request-scoped Logger and log messages.
		l := ctx_glog.Extract(ctx)
		l.Info("some ping")
		l.Info("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}
