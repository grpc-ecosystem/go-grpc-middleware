// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package ratelimit_test

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"google.golang.org/grpc"
)

// alwaysPassLimiter is an example limiter which implements Limiter interface.
// It does not limit any request because Limit function always returns false.
type alwaysPassLimiter struct{}

func (*alwaysPassLimiter) Limit(_ context.Context) error {
	// Example rate limiter could be implemented using e.g. github.com/juju/ratelimit
	//	// Take one token per request. This call doesn't block.
	//	tokenRes := l.tokenBucket.TakeAvailable(1)
	//
	//	// When rate limit reached, return specific error for the clients.
	//	if tokenRes == 0 {
	//		return fmt.Errorf("APP-XXX: reached Rate-Limiting %d", l.tokenBucket.Available())
	//	}
	//
	//	// Rate limit isn't reached.
	//	return nil
	//}
	return nil
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
