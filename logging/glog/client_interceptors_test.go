// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog_test

import (
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/glog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
)

func customClientCodeToLevel(c codes.Code) ctx_glog.Severity {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return ctx_glog.ErrorLevel
	}
	level := grpc_glog.DefaultClientCodeToLevel(c)
	return level
}

func TestGLogClientSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	opts := []grpc_glog.Option{
		grpc_glog.WithLevels(customClientCodeToLevel),
	}
	b := newGLogBaseSuite(t)
	b.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_glog.UnaryClientInterceptor(ctx_glog.NewEntry(b.logger), opts...)),
		grpc.WithStreamInterceptor(grpc_glog.StreamClientInterceptor(ctx_glog.NewEntry(b.logger), opts...)),
	}
	suite.Run(t, &glogClientSuite{b})
}

type glogClientSuite struct {
	*glogBaseSuite
}

func (s *glogClientSuite) TestPing() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "one log statement should be logged")

	assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
	assert.Equal(s.T(), msgs[0]["grpc.method"], "Ping", "all lines must contain the correct method name")
	assert.Equal(s.T(), msgs[0]["msg"], "finished client unary call", "handler's message must contain the correct message")
	assert.Equal(s.T(), msgs[0]["span.kind"], "client", "all lines must contain the kind of call (client)")
	assert.Equal(s.T(), msgs[0]["level"], "debug", "OK codes must be logged on debug level.")

	assert.Contains(s.T(), msgs[0], "grpc.time_ms", "interceptor log statement should contain execution time (duration in ms)")
}

func (s *glogClientSuite) TestPingList() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "one log statement should be logged")

	assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
	assert.Equal(s.T(), msgs[0]["grpc.method"], "PingList", "all lines must contain the correct method name")
	assert.Equal(s.T(), msgs[0]["msg"], "finished client streaming call", "handler's message must contain the correct message")
	assert.Equal(s.T(), msgs[0]["span.kind"], "client", "all lines must contain the kind of call (client)")
	assert.Equal(s.T(), msgs[0]["level"], "debug", "OK codes must be logged on debug level.")
	assert.Contains(s.T(), msgs[0], "grpc.time_ms", "interceptor log statement should contain execution time (duration in ms)")
}

func (s *glogClientSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  codes.Code
		level ctx_glog.Severity
		msg   string
	}{
		{
			code:  codes.Internal,
			level: ctx_glog.WarningLevel,
			msg:   "Internal must remap to ErrorLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.NotFound,
			level: ctx_glog.DebugLevel,
			msg:   "NotFound must remap to InfoLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.FailedPrecondition,
			level: ctx_glog.DebugLevel,
			msg:   "FailedPrecondition must remap to WarnLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.Unauthenticated,
			level: ctx_glog.ErrorLevel,
			msg:   "Unauthenticated is overwritten to ErrorLevel with customClientCodeToLevel override, which probably didn't work",
		},
	} {
		s.SetupTest()
		_, err := s.Client.PingError(
			s.SimpleCtx(),
			&pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})

		assert.Error(s.T(), err, "each call here must return an error")

		msgs := s.getOutputJSONs()
		require.Len(s.T(), msgs, 1, "only a single log message is printed")

		assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
		assert.Equal(s.T(), msgs[0]["grpc.method"], "PingError", "all lines must contain the correct method name")
		assert.Equal(s.T(), msgs[0]["grpc.code"], tcase.code.String(), "all lines must contain a grpc code")
		assert.Equal(s.T(), msgs[0]["level"], tcase.level.String(), tcase.msg)
	}
}

func TestGLogClientOverrideSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skip("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	opts := []grpc_glog.Option{
		grpc_glog.WithDurationField(grpc_glog.DurationToDurationField),
	}
	b := newGLogBaseSuite(t)
	b.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_glog.UnaryClientInterceptor(ctx_glog.NewEntry(b.logger), opts...)),
		grpc.WithStreamInterceptor(grpc_glog.StreamClientInterceptor(ctx_glog.NewEntry(b.logger), opts...)),
	}
	suite.Run(t, &glogClientOverrideSuite{b})
}

type glogClientOverrideSuite struct {
	*glogBaseSuite
}

func (s *glogClientOverrideSuite) TestPing_HasOverrides() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "one log statement should be logged")
	assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
	assert.Equal(s.T(), msgs[0]["grpc.method"], "Ping", "all lines must contain the correct method name")
	assert.Equal(s.T(), msgs[0]["msg"], "finished client unary call", "handler's message must contain the correct message")

	assert.NotContains(s.T(), msgs[0], "grpc.time_ms", "message must not contain default duration")
	assert.Contains(s.T(), msgs[0], "grpc.duration", "message must contain overridden duration")
}

func (s *glogClientOverrideSuite) TestPingList_HasOverrides() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "one log statement should be logged")

	assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
	assert.Equal(s.T(), msgs[0]["grpc.method"], "PingList", "all lines must contain the correct method name")
	assert.Equal(s.T(), msgs[0]["msg"], "finished client streaming call", "log message must be correct")
	assert.Equal(s.T(), msgs[0]["span.kind"], "client", "all lines must contain the kind of call (client)")
	assert.Equal(s.T(), msgs[0]["level"], "debug", "OK codes must be logged on debug level.")

	assert.NotContains(s.T(), msgs[0], "grpc.time_ms", "message must not contain default duration")
	assert.Contains(s.T(), msgs[0], "grpc.duration", "message must contain overridden duration")
}
