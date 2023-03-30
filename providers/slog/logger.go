package slog

import (
	"context"
	"fmt"

	"golang.org/x/exp/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// Compatibility check.
var _ logging.Logger = &Logger{}

func InterceptorLogger(logger *slog.Logger) Logger {
	return Logger{logger}
}

type Logger struct {
	*slog.Logger
}

func (l Logger) With(fields ...string) logging.Logger {
	anys := make([]any, 0, len(fields))
	for _, field := range fields {
		anys = append(anys, any(field))
	}
	return Logger{l.Logger.With(anys...)}
}

func (l Logger) Log(level logging.Level, s string) {
	var lvl slog.Level
	switch level {
	case logging.DEBUG:
		lvl = slog.LevelDebug
	case logging.INFO:
		lvl = slog.LevelInfo
	case logging.WARNING:
		lvl = slog.LevelWarn
	case logging.ERROR:
		lvl = slog.LevelError
	default:
		panic(fmt.Sprintf("slog: unknown level %s", level))
	}
	l.Logger.Log(context.Background(), lvl, s)
}
