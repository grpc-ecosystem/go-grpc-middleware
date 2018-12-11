// Copyright 2018 Zheng Dayu. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_ratelimit

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type limiter interface {
	WaitMaxDuration(time.Duration) bool
}

// UnaryServerInterceptor returns a new unary server interceptors that performs request rate limit.
func UnaryServerInterceptor(rateLimiter limiter, maxWaitDuration time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if ok := rateLimiter.WaitMaxDuration(maxWaitDuration); ok {
			return handler(ctx, req)
		}
		return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleare, please retry later.", info.FullMethod)
	}
}

//
// StreamServerInterceptor returns a new stream server interceptors that performs request rate limit.
func StreamServerInterceptor(rateLimiter limiter, maxWaitDuration time.Duration) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if ok := rateLimiter.WaitMaxDuration(maxWaitDuration); ok {
			return handler(srv, stream)
		}
		return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleare, please retry later.", info.FullMethod)
	}
}
