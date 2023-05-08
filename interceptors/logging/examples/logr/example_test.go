// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logr_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
)

// verbosity https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md#what-method-to-use
const (
	debugVerbosity = 4
	infoVerbosity  = 2
	warnVerbosity  = 1
	errorVerbosity = 0
)

// InterceptorLogger adapts logr logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l logr.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		l := l.WithValues(fields...)
		switch lvl {
		case logging.LevelDebug:
			l.V(debugVerbosity).Info(msg)
		case logging.LevelInfo:
			l.V(infoVerbosity).Info(msg)
		case logging.LevelWarn:
			l.V(warnVerbosity).Info(msg)
		case logging.LevelError:
			l.V(errorVerbosity).Info(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func ExampleInterceptorLogger() {
	logger := klog.NewKlogr()

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

type logrExampleTestSuite struct {
	*testpb.InterceptorTestSuite
	logBuffer *ktesting.BufferTL
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	buffer := &ktesting.BufferTL{}
	cfg := ktesting.NewConfig()
	logger := InterceptorLogger(ktesting.NewLogger(buffer, cfg))

	s := &logrExampleTestSuite{
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

func (s *logrExampleTestSuite) TestPing() {
	ctx := context.Background()
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	logStr := s.logBuffer.String()
	require.Contains(s.T(), logStr, "started call")
	require.Contains(s.T(), logStr, "protocol=\"grpc\"")
	require.Contains(s.T(), logStr, "grpc.component=\"server\"")
	require.Contains(s.T(), logStr, "grpc.service=\"testing.testpb.v1.TestService\"")
	require.Contains(s.T(), logStr, "grpc.method=\"Ping\"")
	require.Contains(s.T(), logStr, "grpc.method_type=\"unary\"")
	require.Contains(s.T(), logStr, "start_time=")
	require.Contains(s.T(), logStr, "grpc.time_ms=")
}
