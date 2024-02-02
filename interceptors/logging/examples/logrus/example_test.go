// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logrus_test

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// InterceptorLogger adapts logrus logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l logrus.FieldLogger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make(map[string]any, len(fields)/2)
		i := logging.Fields(fields).Iterator()
		for i.Next() {
			k, v := i.At()
			f[k] = v
		}
		l := l.WithFields(f)

		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func ExampleInterceptorLogger() {
	logger := logrus.New()

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		// Add any other option (check functions starting with logging.With).
	}

	// You can now create a server with logging instrumentation that e.g. logs when the unary or stream call is started or finished.
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(logger), opts...),
			// Add any other interceptor you want.
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(InterceptorLogger(logger), opts...),
			// Add any other interceptor you want.
		),
	)
	// ...user server.

	// Similarly you can create client that will log for the unary and stream client started or finished calls.
	_, _ = grpc.Dial(
		"some-target",
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(InterceptorLogger(logger), opts...),
			// Add any other interceptor you want.
		),
		grpc.WithChainStreamInterceptor(
			logging.StreamClientInterceptor(InterceptorLogger(logger), opts...),
			// Add any other interceptor you want.
		),
	)
	// Output:
}
