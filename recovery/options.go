// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_recovery

import "golang.org/x/net/context"

var (
	defaultOptions = &options{
		recoveryHandlerFuncContext: nil,
	}
)

type options struct {
	recoveryHandlerFuncContext RecoveryHandlerFuncContext
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// WithRecoveryHandler customizes the function for recovering from a panic.
func WithRecoveryHandler(f RecoveryHandlerFunc) Option {
	return func(o *options) {
		o.recoveryHandlerFuncContext = RecoveryHandlerFuncContext(func(ctx context.Context, p interface{}) error {
			return f(p)
		})
	}
}

// WithRecoveryHandlerContext customizes the function for recovering from a panic.
func WithRecoveryHandlerContext(f RecoveryHandlerFuncContext) Option {
	return func(o *options) {
		o.recoveryHandlerFuncContext = f
	}
}
