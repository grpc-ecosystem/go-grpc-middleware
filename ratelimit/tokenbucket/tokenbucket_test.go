package tokenbucket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenBucketRateLimiter(t *testing.T) {
	l := NewTokenBucketRateLimiter(1*time.Millisecond, 1, 1)
	ok := l.WaitMaxDuration(1 * time.Millisecond)
	assert.True(t, ok)
}
