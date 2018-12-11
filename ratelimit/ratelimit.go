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

type Limiter interface {
	WaitMaxDuration(time.Duration) bool
}

type rateLimiter struct {
	limiter         Limiter
	maxWaitDuration time.Duration
}

func (r *rateLimiter) Wait() bool {
	return r.limiter.WaitMaxDuration(r.maxWaitDuration)
}

type emptyLimiter struct{}

func (e *emptyLimiter) WaitMaxDuration(time.Duration) bool {
	return true
}

func emptyRatelimiter() *rateLimiter {
	return &rateLimiter{
		limiter: &emptyLimiter{},
	}
}

// UnaryServerInterceptor returns a new unary server interceptors that performs request rate limit.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	ratelimiter := emptyRatelimiter()
	for _, opt := range opts {
		opt(ratelimiter)
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if ratelimiter.Wait() {
			return handler(ctx, req)
		}
		return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleare, please retry later.", info.FullMethod)
	}
}

// StreamServerInterceptor returns a new stream server interceptors that performs request rate limit.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	ratelimiter := emptyRatelimiter()
	for _, opt := range opts {
		opt(ratelimiter)
	}
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if ratelimiter.Wait() {
			return handler(srv, stream)
		}
		return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleare, please retry later.", info.FullMethod)
	}
}
