// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package zerolog

import (
	"context"
	"time"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
)

var (
	defaultOptions = &options{
		levelFunc:    nil,
		shouldLog:    grpc_logging.DefaultDeciderMethod,
		codeFunc:     grpc_logging.DefaultErrorToCode,
		durationFunc: DefaultDurationToField,
		messageFunc:  DefaultMessageProducer,
	}
)

type options struct {
	levelFunc    CodeToLevel
	shouldLog    grpc_logging.Decider
	codeFunc     grpc_logging.ErrorToCode
	durationFunc DurationToField
	messageFunc  MessageProducer
}

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultCodeToLevel
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

// Option is the function signature for an Option setter.
type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) zerolog.Level

// DurationToField function defines how to produce duration fields for logging.
type DurationToField func(duration time.Duration) (key string, value interface{})

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log.
func WithDecider(f grpc_logging.Decider) Option {
	return func(o *options) {
		o.shouldLog = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f grpc_logging.ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// WithDurationField customizes the function for mapping request durations to log fields.
func WithDurationField(f DurationToField) Option {
	return func(o *options) {
		o.durationFunc = f
	}
}

// WithMessageProducer customizes the function for message formation.
func WithMessageProducer(f MessageProducer) Option {
	return func(o *options) {
		o.messageFunc = f
	}
}

// DefaultCodeToLevel is the default implementation of gRPC return codes to log levels for server side.
func DefaultCodeToLevel(code codes.Code) zerolog.Level {
	switch code {
	case codes.OK:
		return zerolog.InfoLevel
	case codes.Canceled:
		return zerolog.InfoLevel
	case codes.Unknown:
		return zerolog.ErrorLevel
	case codes.InvalidArgument:
		return zerolog.InfoLevel
	case codes.DeadlineExceeded:
		return zerolog.WarnLevel
	case codes.NotFound:
		return zerolog.InfoLevel
	case codes.AlreadyExists:
		return zerolog.InfoLevel
	case codes.PermissionDenied:
		return zerolog.WarnLevel
	case codes.Unauthenticated:
		return zerolog.InfoLevel
	case codes.ResourceExhausted:
		return zerolog.WarnLevel
	case codes.FailedPrecondition:
		return zerolog.WarnLevel
	case codes.Aborted:
		return zerolog.WarnLevel
	case codes.OutOfRange:
		return zerolog.WarnLevel
	case codes.Unimplemented:
		return zerolog.ErrorLevel
	case codes.Internal:
		return zerolog.ErrorLevel
	case codes.Unavailable:
		return zerolog.WarnLevel
	case codes.DataLoss:
		return zerolog.ErrorLevel
	default:
		return zerolog.ErrorLevel
	}
}

// DefaultClientCodeToLevel is the default implementation of gRPC return codes to log levels for client side.
func DefaultClientCodeToLevel(code codes.Code) zerolog.Level {
	switch code {
	case codes.OK:
		return zerolog.DebugLevel
	case codes.Canceled:
		return zerolog.DebugLevel
	case codes.Unknown:
		return zerolog.InfoLevel
	case codes.InvalidArgument:
		return zerolog.DebugLevel
	case codes.DeadlineExceeded:
		return zerolog.InfoLevel
	case codes.NotFound:
		return zerolog.DebugLevel
	case codes.AlreadyExists:
		return zerolog.DebugLevel
	case codes.PermissionDenied:
		return zerolog.InfoLevel
	case codes.Unauthenticated:
		return zerolog.InfoLevel
	case codes.ResourceExhausted:
		return zerolog.DebugLevel
	case codes.FailedPrecondition:
		return zerolog.DebugLevel
	case codes.Aborted:
		return zerolog.DebugLevel
	case codes.OutOfRange:
		return zerolog.DebugLevel
	case codes.Unimplemented:
		return zerolog.WarnLevel
	case codes.Internal:
		return zerolog.WarnLevel
	case codes.Unavailable:
		return zerolog.WarnLevel
	case codes.DataLoss:
		return zerolog.WarnLevel
	default:
		return zerolog.InfoLevel
	}
}

// DefaultDurationToField is the default implementation of converting request duration to a log field (key and value).
var DefaultDurationToField = DurationToTimeMillisField

// DurationToTimeMillisField converts the duration to milliseconds and uses the key `grpc.time_ms`.
func DurationToTimeMillisField(duration time.Duration) (key string, value interface{}) {
	return "grpc.time_ms", durationToMilliseconds(duration)
}

// DurationToDurationField uses the duration value to log the request duration.
func DurationToDurationField(duration time.Duration) (key string, value interface{}) {
	return "grpc.duration", duration
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()) / 1e6
}

// MessageProducer produces a user defined log message.
type MessageProducer func(ctx context.Context, format string, level zerolog.Level, code codes.Code, err error, fields map[string]interface{})

// DefaultMessageProducer writes the default message.
func DefaultMessageProducer(ctx context.Context, format string, level zerolog.Level, code codes.Code, err error, fields map[string]interface{}) {
	ctxLogger := ctxzerolog.Extract(ctx).Logger()
	event := ctxLogger.WithLevel(level)
	if err != nil {
		event = event.Err(err)
	}
	event.Fields(fields).Msg(format)
}
