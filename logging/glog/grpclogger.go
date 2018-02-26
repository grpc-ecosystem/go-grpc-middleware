// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets glog logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLogger() {
	grpclog.SetLoggerV2(ctx_glog.Logger)
}
