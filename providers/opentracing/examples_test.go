// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentracing_test

import (
	"testing"

	"google.golang.org/grpc"

	grpcopentracing "github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
)

func Example() {
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(grpcopentracing.InterceptorTracer()),
		),
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(grpcopentracing.InterceptorTracer()),
		),
	)

	_, _ = grpc.Dial("",
		grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor(grpcopentracing.InterceptorTracer())),
		grpc.WithStreamInterceptor(tracing.StreamClientInterceptor(grpcopentracing.InterceptorTracer())),
	)
}

func TestExamplesBuildable(t *testing.T) {
	Example()
}
