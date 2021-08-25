// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package tracing

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

type SpanKind string

const (
	SpanKindServer SpanKind = "server"
	SpanKindClient SpanKind = "client"
)

func reportable(tracer Tracer) interceptors.CommonReportableFunc {
	return func(ctx context.Context, c interceptors.CallMeta, isClient bool) (interceptors.Reporter, context.Context) {
		kind := SpanKindServer
		if isClient {
			kind = SpanKindClient
		}

		newCtx, span := tracer.Start(ctx, c.FullMethod(), kind)
		return &reporter{ctx: newCtx, span: span}, newCtx
	}
}

// UnaryClientInterceptor returns a new unary client interceptor that optionally traces the execution of external gRPC calls.
// Tracer will use tags (from tags package) available in current context as fields.
func UnaryClientInterceptor(tracer Tracer) grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(reportable(tracer))
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally traces the execution of external gRPC calls.
// Tracer will use tags (from tags package) available in current context as fields.
func StreamClientInterceptor(tracer Tracer) grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(reportable(tracer))
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally traces endpoint handling.
// Tracer will use tags (from tags package) available in current context as fields.
func UnaryServerInterceptor(tracer Tracer) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(reportable(tracer))
}

// StreamServerInterceptor returns a new stream server interceptors that optionally traces endpoint handling.
// Tracer will use tags (from tags package) available in current context as fields.
func StreamServerInterceptor(tracer Tracer) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(reportable(tracer))
}
