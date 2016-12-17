package grpc_retry

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
)

var (
	DefaultRetriableCodes = []codes.Code{codes.ResourceExhausted, codes.Unavailable}

	defaultOptions = &options{
		max:            0, // disabed
		perCallTimeout: 0, // disabled
		includeHeader:  true,
		codes:          DefaultRetriableCodes,
		backoffFunc:    BackoffLinear(50 * time.Millisecond),
	}
)

// BackoffFunc is a function that returns the backoff time between calls of retry.
type BackoffFunc func(attempt uint) time.Duration

// Disable disables the retry behaviour on this call, or this interceptor.
// The same as WithMax(0)
func Disable() optionsFunc {
	return WithMax(0)
}

// WithMax sets the maximum number of retries on this call, or this interceptor.
func WithMax(maxRetries uint) optionsFunc {
	return func(o *options) {
		o.max = maxRetries
	}
}

// WithBackoff sets the BackoffFunc used to control time between retries.
func WithBackoff(bf BackoffFunc) optionsFunc {
	return func(o *options) {
		o.backoffFunc = bf
	}
}

// WithCodes sets which codes should be retried.
// You cannot autmatically retry on Cancelled and Deadline.
func WithCodes(retryCodes ...codes.Code) optionsFunc {
	return func(o *options) {
		o.codes = retryCodes
	}
}

// WithPerRetryTimeout sets the RPC timeout per call (including initial call) on this call, or this interceptor.
// The context.Deadline of the call takes precedence and sets the maximum time the whole invocation
// will take, but WithPerCallTimeout can be used to limit the RPC time per each call.
// For example, with context.Deadline = now + 10s, and WithPerCallTimeout(3 * time.Seconds), each
// of the retry calls (including the initial one) will have a deadline of now + 3s.
// A value of 0 disables the timeout overrides completely and returns to each retry call using the
// parent context.Deadline.
func WithPerRetryTimeout(timeout time.Duration) optionsFunc {
	return func(o *options) {
		o.perCallTimeout = timeout
	}
}

type options struct {
	max            uint
	perCallTimeout time.Duration
	includeHeader  bool
	codes          []codes.Code
	backoffFunc    BackoffFunc
}

type optionsFunc func(*options)

func applyOptionFuncsOrReuse(opt *options, optFuncs []optionsFunc) *options {
	if len(optFuncs) == 0 {
		return opt
	}
	optCopy := &options{}
	*optCopy = *opt
	for _, f := range optFuncs {
		f(optCopy)
	}
	return optCopy
}

// Context sets per-call options, that allow to override the interceptor defaults.
func Context(ctx context.Context, optFuncs ...optionsFunc) context.Context {
	funcs := append(fromContext(ctx), optFuncs...)
	return context.WithValue(ctx, optionFuncMarker, funcs)
}

func fromContext(ctx context.Context) []optionsFunc {
	funcs, ok := ctx.Value(optionFuncMarker).([]optionsFunc)
	if !ok {
		return emptyOptionsFunc
	}
	return funcs
}
