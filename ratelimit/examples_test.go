package ratelimit_test

import (
	"time"

	"github.com/ceshihao/ratelimiter/tokenbucket"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"google.golang.org/grpc"
)

// Simple example of server initialization code.
func Example() {
	// Create unary/stream rateLimiters, based on token bucket here.
	// You can implement your own ratelimiter for the interface.
	unaryRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 10, 10, 10*time.Second)
	streamRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 5, 5, 5*time.Second)
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			ratelimit.UnaryServerInterceptor(unaryRateLimiter),
		),
		grpc_middleware.WithStreamServerChain(
			ratelimit.StreamServerInterceptor(streamRateLimiter),
		),
	)
}
