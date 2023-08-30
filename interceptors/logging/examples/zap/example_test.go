// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package zap_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
)

// InterceptorLogger adapts zap logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)
		iter := logging.Fields(fields).Iterator()
		for iter.Next() {
			k, v := iter.At()
			f = append(f, zap.Any(k, v))
		}
		l := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

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
	logger := zap.NewExample()

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

type zapExampleTestSuite struct {
	*testpb.InterceptorTestSuite
	observedLogs *observer.ObservedLogs
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	observedZapCore, observedLogs := observer.New(zap.DebugLevel)
	logger := InterceptorLogger(zap.New(observedZapCore))
	s := &zapExampleTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &testpb.TestPingService{},
		},
		observedLogs: observedLogs,
	}

	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(logger)),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(logger)),
	}

	suite.Run(t, s)
}

func (s *zapExampleTestSuite) TestPing() {
	ctx := context.Background()
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	require.Equal(s.T(), 2, s.observedLogs.Len())
	line := s.observedLogs.All()[0]

	contextMap := line.ContextMap()
	require.Equal(s.T(), zap.InfoLevel, line.Level)
	require.Equal(s.T(), "started call", line.Entry.Message)

	require.Equal(s.T(), "Ping", contextMap["grpc.method"])
	require.Equal(s.T(), "grpc", contextMap["protocol"])
	require.Equal(s.T(), "server", contextMap["grpc.component"])

	require.Contains(s.T(), contextMap["peer.address"], "127.0.0.1")
	require.NotEmpty(s.T(), contextMap["grpc.start_time"])
	require.NotEmpty(s.T(), contextMap["grpc.time_ms"])
}
