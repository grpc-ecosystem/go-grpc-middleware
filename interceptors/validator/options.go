// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"context"
)

type options struct {
	shouldFailFast      bool
	onValidationErrFunc OnValidationErr
}
type Option func(*options)

func evaluateOpts(opts []Option) *options {
	optCopy := &options{}
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type OnValidationErr func(ctx context.Context, err error)

// WithOnValidationErrFunc registers function that will be invoked on validation error(s).
func WithOnValidationErrFunc(onValidationErrFunc OnValidationErr) Option {
	return func(o *options) {
		o.onValidationErrFunc = onValidationErrFunc
	}
}

// WithFailFast tells validator to immediately stop doing further validation after first validation error.
// This option is ignored if message is only supporting validator.validatorLegacy interface.
func WithFailFast() Option {
	return func(o *options) {
		o.shouldFailFast = true
	}
}
