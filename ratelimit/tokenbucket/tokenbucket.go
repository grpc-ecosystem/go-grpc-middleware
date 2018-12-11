// Copyright 2018 Zheng Dayu. All Rights Reserved.
// See LICENSE for licensing terms.

package tokenbucket

import (
	"time"

	"github.com/juju/ratelimit"
)

type tokenBucketLimiter struct {
	limiter *ratelimit.Bucket
}

// NewTokenBucketRateLimiter creates a tokenBucketLimiter.
func NewTokenBucketRateLimiter(fillInterval time.Duration, capacity, quantum int64) *tokenBucketLimiter {
	return &tokenBucketLimiter{
		limiter: ratelimit.NewBucketWithQuantum(fillInterval, capacity, quantum),
	}
}

// WaitMaxDuration
func (b *tokenBucketLimiter) WaitMaxDuration(maxWaitDuration time.Duration) bool {
	return b.limiter.WaitMaxDuration(1, maxWaitDuration)
}
