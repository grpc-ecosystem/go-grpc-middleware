package ratelimit_test

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
)

// alwaysPassLimiter is an example limiter which implements Limiter interface.
// It does not limit any request because Limit function always returns false.
type alwaysPassLimiter struct{}

func (*alwaysPassLimiter) Limit(_ context.Context) bool {
	return false
}

// Simple example of server initialization code.
func Example() {
	// Create unary/stream rateLimiters, based on token bucket here.
	// You can implement your own ratelimiter for the interface.
	limiter := &alwaysPassLimiter{}
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			ratelimit.UnaryServerInterceptor(limiter),
		),
		grpc.ChainStreamInterceptor(
			ratelimit.StreamServerInterceptor(limiter),
		),
	)
}
