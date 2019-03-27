// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

var (
	defaultOptions = &options{
		filterOutFunc: nil,
		tracer:        nil,
	}
)

// FilterFunc allows users to provide a function that filters out certain methods from being traced.
//
// If it returns false, the given request will not be traced.
type FilterFunc func(ctx context.Context, fullMethodName string) bool

type options struct {
	filterOutFunc FilterFunc
	tracer        opentracing.Tracer
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

// WithFilterFunc customizes the function used for deciding whether a given call is traced or not.
func WithFilterFunc(f FilterFunc) Option {
	return func(o *options) {
		o.filterOutFunc = f
	}
}

// WithTracer sets a custom tracer to be used for this middleware, otherwise the opentracing.GlobalTracer is used.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(o *options) {
		if tracer == nil {
			o.tracer = opentracing.GlobalTracer()
		} else {
			o.tracer = tracer
		}
	}
}
