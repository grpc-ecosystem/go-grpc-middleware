// See LICENSE for licensing terms.

package ratelimit

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Limiter interface {
	Limit() bool
}

type rateLimiter struct {
	limiter Limiter
}

type emptyLimiter struct{}

func (e *emptyLimiter) Limit() bool {
	return false
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
		if ratelimiter.limiter.Limit() {
			return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleare, please retry later.", info.FullMethod)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that performs rate limiting on the request.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	ratelimiter := emptyRatelimiter()
	fmt.Println(ratelimiter.limiter.Limit())
	for _, opt := range opts {
		opt(ratelimiter)
	}
	fmt.Println(ratelimiter.limiter.Limit())
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		fmt.Println(ratelimiter.limiter.Limit())
		if ratelimiter.limiter.Limit() {
			return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleare, please retry later.", info.FullMethod)
		}
		return handler(srv, stream)
	}
}
