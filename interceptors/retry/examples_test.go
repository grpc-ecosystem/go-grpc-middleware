// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package retry_test

import (
	"context"
	"io"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
)

var cc *grpc.ClientConn

// Simple example of using the default interceptor configuration.
func Example_initialization() {
	_, _ = grpc.Dial("myservice.example.com",
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor()),
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor()),
	)
}

// Complex example with a 100ms linear backoff interval, and retry only on NotFound and Unavailable.
func Example_initializationWithOptions() {
	opts := []retry.CallOption{
		retry.WithBackoff(retry.BackoffLinear(100 * time.Millisecond)),
		retry.WithCodes(codes.NotFound, codes.Aborted),
	}
	_, _ = grpc.Dial("myservice.example.com",
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor(opts...)),
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(opts...)),
	)
}

// Example with an exponential backoff starting with 100ms.
//
// Each next interval is the previous interval multiplied by 2.
func Example_initializationWithExponentialBackoff() {
	opts := []retry.CallOption{
		retry.WithBackoff(retry.BackoffExponential(100 * time.Millisecond)),
	}
	_, _ = grpc.Dial("myservice.example.com",
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor(opts...)),
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(opts...)),
	)
}

// Simple example of an idempotent `ServerStream` call, that will be retried automatically 3 times.
func Example_simpleCall() {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()

	client := testpb.NewTestServiceClient(cc)
	stream, _ := client.PingList(ctx, &testpb.PingListRequest{}, retry.WithMax(3))

	for {
		_, err := stream.Recv() // retries happen here
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}
	}
}

// This is an example of an `Unary` call that will also retry on deadlines.
//
// Because the passed in context has a `5s` timeout, the whole `Ping` invocation should finish
// within that time. However, by default all retried calls will use the parent context for their
// deadlines. This means, that unless you shorten the deadline of each call of the retry, you won't
// be able to retry the first call at all.
//
// `WithPerRetryTimeout` allows you to shorten the deadline of each retry call, allowing you to fit
// multiple retries in the single parent deadline.
func ExampleWithPerRetryTimeout() {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	client := testpb.NewTestServiceClient(cc)
	_, _ = client.Ping(
		ctx,
		&testpb.PingRequest{},
		retry.WithMax(3),
		retry.WithPerRetryTimeout(1*time.Second))
}

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
		out := retry.JitterUp(duration, variance)
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