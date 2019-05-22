package tokenbucket

import (
	"time"

	"github.com/juju/ratelimit"
)

type tokenBucketLimiter struct {
	limiter         *ratelimit.Bucket
	maxWaitDuration time.Duration
}

// NewTokenBucketRateLimiter creates a tokenBucketLimiter.
func NewTokenBucketRateLimiter(fillInterval time.Duration, capacity, quantum int64, maxWaitDuration time.Duration) *tokenBucketLimiter {
	return &tokenBucketLimiter{
		limiter:         ratelimit.NewBucketWithQuantum(fillInterval, capacity, quantum),
		maxWaitDuration: maxWaitDuration,
	}
}

// Limit takes 1 token from the bucket only if it needs to wait for no greater than maxWaitDuration
func (b *tokenBucketLimiter) Limit() bool {
	return !b.limiter.WaitMaxDuration(1, b.maxWaitDuration)
}
