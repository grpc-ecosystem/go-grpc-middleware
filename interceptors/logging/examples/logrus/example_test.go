// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logrus_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

type logrusExampleTestSuite struct {
	*testpb.InterceptorTestSuite
	logBuffer *bytes.Buffer
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	buffer := &bytes.Buffer{}
	testLogger := logrus.New()
	testLogger.Out = buffer
	testLogger.SetFormatter(&logrus.JSONFormatter{})
	logger := InterceptorLogger(testLogger)

	s := &logrusExampleTestSuite{
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

func (s *logrusExampleTestSuite) TestPing() {
	ctx := context.Background()
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	var logMap map[string]interface{}
	err = json.Unmarshal(s.logBuffer.Bytes(), &logMap)
	require.NoError(s.T(), err)

	require.Equal(s.T(), "started call", logMap["msg"])
	require.Equal(s.T(), "grpc", logMap["protocol"])
	require.Equal(s.T(), "server", logMap["grpc.component"])
	require.Equal(s.T(), "testing.testpb.v1.TestService", logMap["grpc.service"])
	require.Equal(s.T(), "Ping", logMap["grpc.method"])
	require.Equal(s.T(), "unary", logMap["grpc.method_type"])
	require.Contains(s.T(), logMap["peer.address"], "127.0.0.1")
	require.NotEmpty(s.T(), logMap["grpc.start_time"])
	require.NotEmpty(s.T(), logMap["grpc.time_ms"])
}
