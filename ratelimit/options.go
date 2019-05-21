// See LICENSE for licensing terms.

package ratelimit

import "time"

const infinityDuration time.Duration = 0x7fffffffffffffff

type Option func(*rateLimiter)

// WithRateLimiter customizes your limiter in the middleware
func WithRateLimiter(l Limiter) Option {
	return func(r *rateLimiter) {
		r.limiter = l
	}
}
