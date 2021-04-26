// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logrus

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// Compatibility check.
var _ logging.Logger = &Logger{}

// Logger is a logrus logging adapter compatible with logging middlewares.
type Logger struct {
	logrus.FieldLogger
}

// InterceptorLogger converts logrus logger to Logger adapter.
func InterceptorLogger(logger logrus.FieldLogger) *Logger {
	return &Logger{logger}
}

// Log implements logging.Logger interface.
func (l *Logger) Log(lvl logging.Level, msg string) {
	switch lvl {
	case logging.DEBUG:
		l.Debug(msg)
	case logging.INFO:
		l.Info(msg)
	case logging.WARNING:
		l.Warn(msg)
	case logging.ERROR:
		l.Error(msg)
	default:
		panic(fmt.Sprintf("logrus: unknown level %s", lvl))
	}
}

// With implements logging.Logger interface.
func (l *Logger) With(fields ...string) logging.Logger {
	vals := make(map[string]interface{}, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		vals[fields[i]] = fields[i+1]
	}
	return InterceptorLogger(l.WithFields(vals))
}
