// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logrus_test

import "github.com/mwitkow/go-grpc-middleware/logging/logrus"
import (
	"github.com/Sirupsen/logrus"
	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"
	"golang.org/x/net/context"
)

// Simple unary handler that adds custom fields to the requests's context
func Example_ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_logrus.AddFields(ctx, logrus.Fields{"custom_string": "something", "custom_int": 1337})
	grpc_logrus.Extract(ctx).Info("some ping")
	return &pb_testproto.PingResponse{Value: ping.Value}, nil
}
