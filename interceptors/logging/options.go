package logging

import (
	"fmt"
	"time"
)

var (
	defaultOptions = &options{
		shouldLog:         DefaultDeciderMethod,
		codeFunc:          DefaultErrorToCode,
		durationFieldFunc: DefaultDurationToFields,
		// levelFunc depends if it's client or server.
		levelFunc: nil,
		// request logging is switched off by default.
		shouldLogRequest: DefaultRequestDecider,
	}
)

type options struct {
	levelFunc         CodeToLevel
	shouldLog         Decider
	codeFunc          ErrorToCode
	durationFieldFunc DurationToFields
	shouldLogRequest  RequestDecider
}

type Option func(*options)

// DurationToFields function defines how to produce duration fields for logging.
type DurationToFields func(duration time.Duration) Fields

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultServerCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateClientOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultClientCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log.
func WithDecider(f Decider) Option {
	return func(o *options) {
		o.shouldLog = f
	}
}

// WithRequestLoggingDecider customizes the function for deciding if the gRPC interceptor logs should log the request details.
func WithRequestLoggingDecider(f RequestDecider) Option {
	return func(o *options) {
		o.shouldLogRequest = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// WithDurationField customizes the function for mapping request durations to log fields.
func WithDurationField(f DurationToFields) Option {
	return func(o *options) {
		o.durationFieldFunc = f
	}
}

// DefaultDurationToFields is the default implementation of converting request duration to a field.
var DefaultDurationToFields = DurationToTimeMillisFields

// DurationToTimeMillisFields converts the duration to milliseconds and uses the key `grpc.time_ms`.
func DurationToTimeMillisFields(duration time.Duration) Fields {
	return Fields{"grpc.time_ms", fmt.Sprintf("%v", durationToMilliseconds(duration))}
}

// DurationToDurationField uses a Duration field to log the request duration
// and leaves it up to Log's encoder settings to determine how that is output.
func DurationToDurationField(duration time.Duration) Fields {
	return Fields{"grpc.duration", duration.String()}
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}

func DefaultRequestDecider(_ string, _ error) bool {
	return false
}
