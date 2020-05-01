package kit

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

type Logger struct {
	log.Logger
}

func InterceptorLogger(logger log.Logger) *Logger {
	return &Logger{logger}
}

func (l *Logger) Log(lvl logging.Level, msg string) {
	switch lvl {
	case logging.DEBUG:
		_ = level.Debug(l.Logger).Log(msg)
	case logging.INFO:
		_ = level.Info(l.Logger).Log(msg)
	case logging.WARNING:
		_ = level.Warn(l.Logger).Log(msg)
	case logging.ERROR:
		_ = level.Error(l.Logger).Log(msg)
	default:
		panic(fmt.Sprintf("kit: unknown level %s", lvl))
	}
}

func (l *Logger) With(fields ...string) logging.Logger {
	vals := make([]interface{}, 0, len(fields))
	for _, v := range fields {
		vals = append(vals, v)
	}
	return InterceptorLogger(log.With(l.Logger, vals...))
}
