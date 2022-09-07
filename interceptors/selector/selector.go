// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package selector

import (
	"context"

	"google.golang.org/grpc"
)

type MatchFunc func(ctx context.Context, fullMethod string) bool

// UnaryServerInterceptor returns a new unary server interceptor that will decide whether to call
// the interceptor on the behavior of the MatchFunc.
func UnaryServerInterceptor(interceptors grpc.UnaryServerInterceptor, match MatchFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if match(ctx, info.FullMethod) {
			return interceptors(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that will decide whether to call
// the interceptor on the behavior of the MatchFunc.
func StreamServerInterceptor(interceptors grpc.StreamServerInterceptor, match MatchFunc) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if match(ss.Context(), info.FullMethod) {
			return interceptors(srv, ss, info, handler)
		}
		return handler(srv, ss)
	}
}
