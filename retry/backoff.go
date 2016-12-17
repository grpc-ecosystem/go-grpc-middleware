// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_retry

import "time"

// BackoffLinear is very simple: it waits for a fixed period of time between calls.
func BackoffLinear(waitBetween time.Duration) BackoffFunc {
	return func(attempt uint) time.Duration {
		return waitBetween
	}
}