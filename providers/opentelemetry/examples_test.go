// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentelemetry_test

import (
	"testing"

	"google.golang.org/grpc"

	grpcopentelemetry "github.com/grpc-ecosystem/go-grpc-middleware/providers/opentelemetry/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
)

func Example() {
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(grpcopentelemetry.InterceptorTracer()),
		),
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(grpcopentelemetry.InterceptorTracer()),
		),
	)

	_, _ = grpc.Dial("",
		grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor(grpcopentelemetry.InterceptorTracer())),
		grpc.WithStreamInterceptor(tracing.StreamClientInterceptor(grpcopentelemetry.InterceptorTracer())),
	)
}

func TestExamplesBuildable(t *testing.T) {
	Example()
}
