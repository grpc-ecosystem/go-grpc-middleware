// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog

import (
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	"google.golang.org/grpc/codes"
)

var (
	defaultOptions = &options{
		levelFunc:    nil,
		shouldLog:    grpc_logging.DefaultDeciderMethod,
		codeFunc:     grpc_logging.DefaultErrorToCode,
		durationFunc: DefaultDurationToField,
	}
)

type options struct {
	levelFunc    CodeToLevel
	shouldLog    grpc_logging.Decider
	codeFunc     grpc_logging.ErrorToCode
	durationFunc DurationToField
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

type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) ctx_glog.Severity

// DurationToField function defines how to produce duration fields for logging
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

// DefaultCodeToLevel is the default implementation of gRPC return codes to log levels for server side.
func DefaultCodeToLevel(code codes.Code) ctx_glog.Severity {
	switch code {
	case codes.OK:
		return ctx_glog.InfoLevel
	case codes.Canceled:
		return ctx_glog.InfoLevel
	case codes.Unknown:
		return ctx_glog.ErrorLevel
	case codes.InvalidArgument:
		return ctx_glog.InfoLevel
	case codes.DeadlineExceeded:
		return ctx_glog.WarningLevel
	case codes.NotFound:
		return ctx_glog.InfoLevel
	case codes.AlreadyExists:
		return ctx_glog.InfoLevel
	case codes.PermissionDenied:
		return ctx_glog.WarningLevel
	case codes.Unauthenticated:
		return ctx_glog.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return ctx_glog.WarningLevel
	case codes.FailedPrecondition:
		return ctx_glog.WarningLevel
	case codes.Aborted:
		return ctx_glog.WarningLevel
	case codes.OutOfRange:
		return ctx_glog.WarningLevel
	case codes.Unimplemented:
		return ctx_glog.ErrorLevel
	case codes.Internal:
		return ctx_glog.ErrorLevel
	case codes.Unavailable:
		return ctx_glog.WarningLevel
	case codes.DataLoss:
		return ctx_glog.ErrorLevel
	default:
		return ctx_glog.ErrorLevel
	}
}

// DefaultClientCodeToLevel is the default implementation of gRPC return codes to log levels for client side.
func DefaultClientCodeToLevel(code codes.Code) ctx_glog.Severity {
	switch code {
	case codes.OK:
		return ctx_glog.DebugLevel
	case codes.Canceled:
		return ctx_glog.DebugLevel
	case codes.Unknown:
		return ctx_glog.InfoLevel
	case codes.InvalidArgument:
		return ctx_glog.DebugLevel
	case codes.DeadlineExceeded:
		return ctx_glog.InfoLevel
	case codes.NotFound:
		return ctx_glog.DebugLevel
	case codes.AlreadyExists:
		return ctx_glog.DebugLevel
	case codes.PermissionDenied:
		return ctx_glog.InfoLevel
	case codes.Unauthenticated:
		return ctx_glog.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return ctx_glog.DebugLevel
	case codes.FailedPrecondition:
		return ctx_glog.DebugLevel
	case codes.Aborted:
		return ctx_glog.DebugLevel
	case codes.OutOfRange:
		return ctx_glog.DebugLevel
	case codes.Unimplemented:
		return ctx_glog.WarningLevel
	case codes.Internal:
		return ctx_glog.WarningLevel
	case codes.Unavailable:
		return ctx_glog.WarningLevel
	case codes.DataLoss:
		return ctx_glog.WarningLevel
	default:
		return ctx_glog.InfoLevel
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
	return float32(duration.Nanoseconds()/1000) / 1000
}
