// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package slog_test

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func ExampleInterceptorLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stderr))

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

type slogExampleTestSuite struct {
	*testpb.InterceptorTestSuite
	logBuffer *bytes.Buffer
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	buffer := &bytes.Buffer{}
	logger := InterceptorLogger(slog.New(slog.NewTextHandler(buffer)))
	s := &slogExampleTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &testpb.TestPingService{},
		},
		logBuffer: buffer,
	}

	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(logger, logging.WithLogOnEvents(logging.StartCall))),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(logger, logging.WithLogOnEvents(logging.StartCall))),
	}

	suite.Run(t, s)
}

func (s *slogExampleTestSuite) TestPing() {
	ctx := context.Background()
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	logStr := s.logBuffer.String()
	require.Contains(s.T(), logStr, "started call")
	require.Contains(s.T(), logStr, "protocol=grpc")
	require.Contains(s.T(), logStr, "grpc.component=server")
	require.Contains(s.T(), logStr, "grpc.service=testing.testpb.v1.TestService")
	require.Contains(s.T(), logStr, "grpc.method=Ping")
	require.Contains(s.T(), logStr, "grpc.method_type=unary")
	require.Contains(s.T(), logStr, "start_time=")
	require.Contains(s.T(), logStr, "grpc.time_ms=")
}
