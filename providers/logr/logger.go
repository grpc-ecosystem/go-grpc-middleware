// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logr

import (
	"fmt"

	"github.com/go-logr/logr"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// verbosity https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md#what-method-to-use
const (
	debugVerbosity = 4
	infoVerbosity  = 2
	warnVerbosity  = 1
	errorVerbosity = 0
)

// Compatibility check.
var _ logging.Logger = &Logger{}

// Logger is a logr logging adapter compatible with logging middlewares.
type Logger struct {
	logr.Logger
}

// InterceptorLogger converts zap logger to Logger adapter.
func InterceptorLogger(logger logr.Logger) *Logger {
	return &Logger{Logger: logger}
}

// Log implements logging.Logger interface.
func (l *Logger) Log(lvl logging.Level, msg string) {
	switch lvl {
	case logging.DEBUG:
		l.Logger.V(debugVerbosity).Info(msg)
	case logging.INFO:
		l.Logger.V(infoVerbosity).Info(msg)
	case logging.WARNING:
		l.Logger.V(warnVerbosity).Info(msg)
	case logging.ERROR:
		l.Logger.V(errorVerbosity).Info(msg)
	default:
		panic(fmt.Sprintf("logr: unknown level %s", lvl))
	}
}

// With implements logging.Logger interface.
func (l *Logger) With(fields ...string) logging.Logger {
	vals := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		vals[i] = fields[i]
	}

	return InterceptorLogger(l.Logger.WithValues(vals...))
}
