// Copyright 2018 Zheng Dayu. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_ratelimit

import "time"

const infinityDuration time.Duration = 0x7fffffffffffffff

type Option func(*rateLimiter)

// WithLimiter customizes your limiter in the middleware
func WithLimiter(l Limiter) Option {
	return func(r *rateLimiter) {
		r.limiter = l
		if r.maxWaitDuration == 0 {
			r.maxWaitDuration = infinityDuration
		}
	}
}

// WithMaxWaitDuration customizes maxWaitDuration in limiter's WaitMaxDuration action.
func WithMaxWaitDuration(maxWaitDuration time.Duration) Option {
	return func(r *rateLimiter) {
		r.maxWaitDuration = maxWaitDuration
	}
}
