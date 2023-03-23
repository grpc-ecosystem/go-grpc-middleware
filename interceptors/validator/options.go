package validator

import "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

var (
	defaultOptions = &options{
		level:          "",
		logger:         nil,
		shouldFailFast: false,
	}
)

type options struct {
	level          logging.Level
	logger         logging.Logger
	shouldFailFast bool
}

type Option func(*options)

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateClientOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// WithLogger tells validator to log all the error.
func WithLogger(level logging.Level, logger logging.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithFailFast tells validator to immediately stop doing further validation after first validation error.
func WithFailFast() Option {
	return func(o *options) {
		o.shouldFailFast = true
	}
}
