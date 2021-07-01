package grpc_zap

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/grpclog"
)

func Test_zapGrpcLogger_V(t *testing.T) {
	// copied from gRPC
	const (
		// infoLog indicates Info severity.
		infoLog int = iota
		// warningLog indicates Warning severity.
		warningLog
		// errorLog indicates Error severity.
		errorLog
		// fatalLog indicates Fatal severity.
		fatalLog
	)

	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	ReplaceGrpcLoggerV2WithVerbosity(logger, warningLog)
	assert.False(t, grpclog.V(infoLog))
	assert.True(t, grpclog.V(warningLog))
	assert.True(t, grpclog.V(errorLog))
	assert.True(t, grpclog.V(fatalLog))
}

func TestReplaceGrpcLoggerV2(t *testing.T) {
	defer ReplaceGrpcLoggerV2(zap.NewNop())

	args := []interface{}{"message", "param"}
	cases := []struct {
		name  string
		fn    func(...interface{})
		level zapcore.Level
	}{
		{name: "Info", fn: grpclog.Info, level: zap.InfoLevel},
		{name: "Infoln", fn: grpclog.Infoln, level: zap.InfoLevel},
		{name: "Warning", fn: grpclog.Warning, level: zap.WarnLevel},
		{name: "Warningln", fn: grpclog.Warningln, level: zap.WarnLevel},
		{name: "Error", fn: grpclog.Error, level: zap.ErrorLevel},
		{name: "Errorln", fn: grpclog.Errorln, level: zap.ErrorLevel},
		{name: "Fatal", fn: grpclog.Fatal, level: zap.FatalLevel},
		{name: "Fatalln", fn: grpclog.Fatalln, level: zap.FatalLevel},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			called := false
			ReplaceGrpcLoggerV2(zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
				called = true
				assert.Equal(t, c.level, entry.Level)
				assert.Equal(t, fmt.Sprint(args...), entry.Message)
				_, file, _, _ := runtime.Caller(0)
				assert.Equal(t, file, entry.Caller.File)
				return nil
			}), zap.AddCaller(), zap.OnFatal(zapcore.WriteThenPanic))))

			if c.level != zap.FatalLevel {
				c.fn(args...)
			} else {
				assert.Panics(t, func() {
					c.fn(args...)
				})
			}
			assert.True(t, called, "hook not called")
		})
	}
}
