// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentracing

import (
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/grpclog"
)

var (
	defaultOptions = &options{
		tracer: nil,
	}
)

type options struct {
	tracer          opentracing.Tracer
	traceHeaderName string
	errorLogFunc    ErrorLogFunc
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	if optCopy.tracer == nil {
		optCopy.tracer = opentracing.GlobalTracer()
	}
	if optCopy.traceHeaderName == "" {
		optCopy.traceHeaderName = "uber-trace-id"
	}
	if optCopy.errorLogFunc == nil {
		optCopy.errorLogFunc = grpclog.Infof
	}
	return optCopy
}

type Option func(*options)

// ErrorLogFunc is a func that log grpc_opentracing errors if not provided will be used grpclog.Infof
type ErrorLogFunc func(format string, args ...interface{})

// WithTraceHeaderName customizes the trace header name where trace metadata passed with requests.
// Default one is `uber-trace-id`
func WithTraceHeaderName(name string) Option {
	return func(o *options) {
		o.traceHeaderName = name
	}
}

// WithTracer sets a custom tracer to be used for this middleware, otherwise the opentracing.GlobalTracer is used.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}

// WithErrorLogFunc customizes logging grpc_opentracing errors
func WithErrorLogFunc(f ErrorLogFunc) Option {
	return func(o *options) {
		o.errorLogFunc = f
	}
}
