// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zap_test

import (
	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"

	"context"

	"github.com/mwitkow/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
)

// Simple unary handler that adds custom fields to the requests's context
func Example_ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_zap.AddFields(ctx, zap.String("custom_string", "something"), zap.Int("custom_int", 1337))
	grpc_zap.Extract(ctx).Info("some ping")
	return &pb_testproto.PingResponse{Value: ping.Value}, nil
}
