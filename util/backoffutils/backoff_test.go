// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package backoffutils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/util/backoffutils"
)

// scale duration by a factor
func scaleDuration(d time.Duration, factor float64) time.Duration {
	return time.Duration(float64(d) * factor)
}

func TestJitterUp(t *testing.T) {
	// arguments to jitterup
	duration := 10 * time.Second
	variance := 0.10

	// bound to check
	max := 11000 * time.Millisecond
	min := 9000 * time.Millisecond
	high := scaleDuration(max, 0.98)
	low := scaleDuration(min, 1.02)

	highCount := 0
	lowCount := 0

	for i := 0; i < 1000; i++ {
		out := backoffutils.JitterUp(duration, variance)
		assert.True(t, out <= max, "value %s must be <= %s", out, max)
		assert.True(t, out >= min, "value %s must be >= %s", out, min)

		if out > high {
			highCount++
		}
		if out < low {
			lowCount++
		}
	}

	assert.True(t, highCount != 0, "at least one sample should reach to >%s", high)
	assert.True(t, lowCount != 0, "at least one sample should to <%s", low)
}
