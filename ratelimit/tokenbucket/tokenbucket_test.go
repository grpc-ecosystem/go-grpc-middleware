package tokenbucket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenBucketRateLimiter_LimitPass(t *testing.T) {
	l := NewTokenBucketRateLimiter(1*time.Millisecond, 1, 1, 1*time.Millisecond)
	ok := l.Limit()
	assert.False(t, ok)
}

func TestTokenBucketRateLimiter_LimitFail(t *testing.T) {
	l := NewTokenBucketRateLimiter(10*time.Second, 1, 1, 1*time.Millisecond)
	ok := l.Limit()
	assert.False(t, ok)
	ok2 := l.Limit()
	assert.True(t, ok2)
}
