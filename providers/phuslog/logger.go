// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package phuslog

import (
	"fmt"

	"github.com/phuslu/log"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// Compatibility check.
var _ logging.Logger = &Logger{}

// Logger is a phuslog logging adapter compatible with logging middlewares.
type Logger struct {
	log.GrcpGatewayLogger
}

// InterceptorLogger is a phuslog.Logger to Logger adapter.
func InterceptorLogger(logger log.GrcpGatewayLogger) *Logger {
	return &Logger{logger}
}

// Log implements the logging.Logger interface.
func (l *Logger) Log(lvl logging.Level, msg string) {
	switch lvl {
	case logging.DEBUG:
		l.Debug(msg)
	case logging.INFO:
		l.Info(msg)
	case logging.WARNING:
		l.Warning(msg)
	case logging.ERROR:
		l.Error(msg)
	default:
		l.With("error-lvl", fmt.Sprintf("phuslog: unknown level %s", lvl)).Log(logging.ERROR, msg)
	}
}

// With implements the logging.Logger interface.
func (l *Logger) With(fields ...string) logging.Logger {
	vals := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		vals[i] = fields[i]
	}
	return InterceptorLogger(l.WithValues(vals...))
}
