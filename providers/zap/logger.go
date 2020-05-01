package zap

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

type Logger struct {
	*zap.Logger
}

func InterceptorLogger(logger *zap.Logger) *Logger {
	return &Logger{logger}
}

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
		panic(fmt.Sprintf("zap: unknown level %s", lvl))
	}
}

func (l *Logger) With(fields ...string) logging.Logger {
	vals := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		vals = append(vals, zap.String(fields[i], fields[i+1]))
	}
	return InterceptorLogger(l.Logger.With(vals...))
}
