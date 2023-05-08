// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package log_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

// InterceptorLogger adapts standard Go logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *log.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			msg = fmt.Sprintf("DEBUG :%v", msg)
		case logging.LevelInfo:
			msg = fmt.Sprintf("INFO :%v", msg)
		case logging.LevelWarn:
			msg = fmt.Sprintf("WARN :%v", msg)
		case logging.LevelError:
			msg = fmt.Sprintf("ERROR :%v", msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
		l.Println(append([]any{"msg", msg}, fields...))
	})
}

func ExampleInterceptorLogger() {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

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

type logExampleTestSuite struct {
	*testpb.InterceptorTestSuite
	logBuffer *bytes.Buffer
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	buffer := &bytes.Buffer{}
	logger := InterceptorLogger(log.New(buffer, "", 0))

	s := &logExampleTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &testpb.TestPingService{},
		},
		logBuffer: buffer,
	}

	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(logger)),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(logger)),
	}

	suite.Run(t, s)
}

func (s *logExampleTestSuite) TestPing() {
	ctx := context.Background()
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	logStr := s.logBuffer.String()
	require.Contains(s.T(), logStr, "msg INFO")
	require.Contains(s.T(), logStr, ":started call")
	require.Contains(s.T(), logStr, "protocol grpc")
	require.Contains(s.T(), logStr, "grpc.component server")
	require.Contains(s.T(), logStr, "grpc.service testing.testpb.v1.TestService")
	require.Contains(s.T(), logStr, "grpc.method Ping")
	require.Contains(s.T(), logStr, "grpc.method_type unary")
	require.Contains(s.T(), logStr, "start_time ")
	require.Contains(s.T(), logStr, "grpc.time_ms")

}
