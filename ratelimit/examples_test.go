// Copyright 2018 Zheng Dayu. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_ratelimit_test

import (
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit/tokenbucket"
	"google.golang.org/grpc"
)

// Simple example of server initialization code.
func Example_initialization() {
	// Create unary/stream rateLimiters, based on token bucket here.
	// You can implement your own ratelimiter for the interface.
	unaryRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 10, 10)
	streamRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 5, 5)
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ratelimit.UnaryServerInterceptor(unaryRateLimiter, 1*time.Second),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ratelimit.StreamServerInterceptor(streamRateLimiter, 10*time.Second),
		),
	)
}
