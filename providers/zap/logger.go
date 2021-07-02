package zap

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// Compatibility check.
var _ logging.Logger = &Logger{}

// Logger is a zap logging adapter compatible with logging middlewares.
type Logger struct {
	*zap.Logger
}

// InterceptorLogger converts zap logger to Logger adapter.
func InterceptorLogger(logger *zap.Logger) *Logger {
	return &Logger{logger}
}

// Log implements logging.Logger interface.
func (l *Logger) Log(lvl logging.Level, msg string) {
	switch lvl {
	case logging.DEBUG:
		l.WithOptions(zap.AddCallerSkip(1)).Debug(msg)
	case logging.INFO:
		l.WithOptions(zap.AddCallerSkip(1)).Info(msg)
	case logging.WARNING:
		l.WithOptions(zap.AddCallerSkip(1)).Warn(msg)
	case logging.ERROR:
		l.WithOptions(zap.AddCallerSkip(1)).Error(msg)
	default:
		panic(fmt.Sprintf("zap: unknown level %s", lvl))
	}
}

// With implements logging.Logger interface.
func (l *Logger) With(fields ...string) logging.Logger {
	vals := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		vals = append(vals, zap.String(fields[i], fields[i+1]))
	}
	return InterceptorLogger(l.Logger.With(vals...))
}
