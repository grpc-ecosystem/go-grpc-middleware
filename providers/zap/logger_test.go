// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package zap

import (
	"runtime"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

func TestLogger_Log(t *testing.T) {
	msg := "message"

	levels := []logging.Level{logging.DEBUG, logging.INFO, logging.WARNING, logging.ERROR}
	for _, level := range levels {
		called := false
		logger := InterceptorLogger(zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
			called = true

			if entry.Message != msg {
				t.Errorf("expect %v, got %v", msg, entry.Message)
			}
			if _, file, _, _ := runtime.Caller(0); entry.Caller.File != file {
				t.Errorf("caller: expected %v, got %v", file, entry.Caller.File)
			}
			return nil
		}), zap.AddCaller())))

		logger.Log(level, msg)
		if !called {
			t.Error("hook isn't called")
		}
	}
}
