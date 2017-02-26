// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package backoffutils_test

import (
	"testing"
	"time"

	"github.com/mwitkow/go-grpc-middleware/util/backoffutils"
	"github.com/stretchr/testify/assert"
)

func TestJitterUp(t *testing.T) {
	for i := 0; i < 1000; i++ {
		out := backoffutils.JitterUp(10*time.Second, 0.10)
		assert.True(t, out <= 11*time.Second, "value must be <= 11s")
		assert.True(t, out >= 9*time.Second, "value must be >= 9s")
	}
}
