package tokenbucket

// Implement Limiter interface.

import (
	"context"
	"fmt"

	"github.com/juju/ratelimit"
)

// TockenBucketInterceptor implement tocken bucket algorithm.
type TockenBucketInterceptor struct {
	tokenBucket *ratelimit.Bucket
}

// Limit Implement Limiter interface
func (r *TokenBucketInterceptor) Limit(_ context.Context) error {
	// Take one token pro request. This call doesn't block.
	tokenRes := r.tokenBucket.TakeAvailable(1)

	// When rate limit reached, return specific error for the clients.
	if tokenRes == 0 {
		return fmt.Errorf("APP-XXX: reached Rate-Limiting %d", r.tokenBucket.Available())
	}

	// Rate limit isn't reached.
	return nil
}
