// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.
package retry

import (
	"context"
	"testing"
	"time"
)

func TestBackoffExponentialWithJitter(t *testing.T) {
	scalar := 100 * time.Millisecond
	jitterFrac := 0.10
	backoffFunc := BackoffExponentialWithJitter(scalar, jitterFrac)
	// use 64 so we are past number of attempts where exponentBase2 would overflow
	for i := 0; i < 64; i++ {
		waitFor := backoffFunc(nil, uint(i))
		if waitFor < 0 {
			t.Errorf("BackoffExponentialWithJitter(%d) = %d; want >= 0", i, waitFor)
		}
	}
}

func TestBackoffExponentialWithJitterBounded(t *testing.T) {
	scalar := 100 * time.Millisecond
	jitterFrac := 0.10
	maxBound := 10 * time.Second
	backoff := BackoffExponentialWithJitterBounded(scalar, jitterFrac, maxBound)
	// use 64 so we are past number of attempts where exponentBase2 would overflow
	for i := 0; i < 64; i++ {
		waitFor := backoff(context.Background(), uint(i))
		if waitFor > maxBound {
			t.Fatalf("expected dur to be less than %v, got %v for %d", maxBound, waitFor, i)
		}
		if waitFor < 0 {
			t.Fatalf("expected dur to be greater than 0, got %v for %d", waitFor, i)
		}
	}
}
