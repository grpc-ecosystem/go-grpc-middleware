// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	"golang.org/x/net/context"
)

// AddFields adds glog fields to the logger.
// Deprecated: should use the ctx_glog.Extract instead
func AddFields(ctx context.Context, fields ctx_glog.Fields) {
	ctx_glog.AddFields(ctx, fields)
}

// Extract takes the call-scoped ctx_glog.Entry from grpc_glog middleware.
// Deprecated: should use the ctx_glog.Extract instead
func Extract(ctx context.Context) *ctx_glog.Entry {
	return ctx_glog.Extract(ctx)
}
