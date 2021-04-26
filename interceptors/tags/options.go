// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package tags

var (
	defaultOptions = &options{
		requestFieldsFunc: nil,
	}
)

type options struct {
	requestFieldsFunc RequestFieldExtractorFunc
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

// WithFieldExtractor customizes the function for extracting log fields from protobuf messages, for
// unary and server-streamed methods only.
func WithFieldExtractor(f RequestFieldExtractorFunc) Option {
	return func(o *options) {
		o.requestFieldsFunc = f
	}
}
