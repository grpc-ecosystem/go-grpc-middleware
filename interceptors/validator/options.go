package validator

import "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

var (
	defaultOptions = &options{
		logger:         DefaultLoggerMethod,
		shouldFailFast: DefaultDeciderMethod,
	}
)

type options struct {
	logger         Logger
	shouldFailFast Decider
}

// Option
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

// Logger
type Logger func() (logging.Level, logging.Logger)

// DefaultLoggerMethod
func DefaultLoggerMethod() (logging.Level, logging.Logger) {
	return "", nil
}

// WithLogger
func WithLogger(logger Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// Decision
type Decision bool

// Decider function defines rules for suppressing any interceptor logs.
type Decider func() Decision

// DefaultDeciderMethod
func DefaultDeciderMethod() Decision {
	return false
}

// WithFailFast
func WithFailFast(d Decider) Option {
	return func(o *options) {
		o.shouldFailFast = d
	}
}
