// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logrus_test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"bytes"

	"encoding/json"
	"io"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-grpc-middleware/logging"
	"github.com/mwitkow/go-grpc-middleware/logging/logrus"
	"github.com/mwitkow/go-grpc-middleware/testing"
	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"strings"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type loggingPingService struct {
	pb_testproto.TestServiceServer
}

func customCodeToLevel(c codes.Code) logrus.Level {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return logrus.ErrorLevel
	}
	level := grpc_logrus.DefaultCodeToLevel(c)
	return level
}

func (s *loggingPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_logrus.AddFields(ctx, logrus.Fields{"custom_string": "something", "custom_int": 1337})
	grpc_logrus.Extract(ctx).Info("some ping")
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *loggingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *loggingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	grpc_logrus.AddFields(stream.Context(), logrus.Fields{"custom_string": "something", "custom_int": 1337})
	grpc_logrus.Extract(stream.Context()).Info("some pinglist")
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *loggingPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	// This hijacks the PingEmpty to test whether the given interceptor implements the grpc_logging metadata.
	grpc_logging.ExtractMetadata(ctx).AddFieldsFromMiddleware(
		[]string{"middleware_1", "middleware_2"},
		[]interface{}{1410, "some_content"})
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

func TestLogrusLoggingSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	b := &bytes.Buffer{}
	log := logrus.New()
	log.Out = b
	log.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	opts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(customCodeToLevel),
	}
	s := &LogrusLoggingSuite{
		log:    logrus.NewEntry(log),
		buffer: b,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &loggingPingService{&grpc_testing.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_logrus.StreamServerInterceptor(logrus.NewEntry(log), opts...)),
				grpc.UnaryInterceptor(grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(log), opts...)),
			},
		},
	}
	suite.Run(t, s)
}

type LogrusLoggingSuite struct {
	*grpc_testing.InterceptorTestSuite
	buffer *bytes.Buffer
	log    *logrus.Entry
}

func (s *LogrusLoggingSuite) SetupTest() {
	s.buffer.Reset()
}

func (s *LogrusLoggingSuite) getOutputJSONs() []string {
	ret := []string{}
	dec := json.NewDecoder(s.buffer)
	for {
		var val map[string]json.RawMessage
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.T().Fatalf("failed decoding output from Logrus JSON: %v", err)
		}
		out, _ := json.MarshalIndent(val, "", "  ")
		ret = append(ret, string(out))
	}
	return ret
}

func (s *LogrusLoggingSuite) TestPing_WithCustomTags() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	msgs := s.getOutputJSONs()
	assert.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		s.T()
		assert.Contains(s.T(), m, `"grpc_service": "mwitkow.testproto.TestService"`, "all lines must contain service name")
		assert.Contains(s.T(), m, `"grpc_method": "Ping"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"custom_string": "something"`, "all lines must contain `custom_string` set by AddFields")
		assert.Contains(s.T(), m, `"custom_int": 1337`, "all lines must contain `custom_int` set by AddFields")
		// request field extraction
		assert.Contains(s.T(), m, `"request.value": "something"`, "all lines must contain fields extracted from goodPing because of test.manual_extractfields.pb")
	}
	assert.Contains(s.T(), msgs[0], `"msg": "some ping"`, "handler's message must contain user message")
	assert.Contains(s.T(), msgs[1], `"msg": "finished unary call"`, "interceptor message must contain string")
	assert.Contains(s.T(), msgs[1], `"level": "info"`, "OK error codes must be logged on info level.")
	assert.Contains(s.T(), msgs[1], `"grpc_time_ns":`, "interceptor log statement should contain execution time")
}

func (s *LogrusLoggingSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  codes.Code
		level logrus.Level
		msg   string
	}{
		{
			code:  codes.Internal,
			level: logrus.ErrorLevel,
			msg:   "Internal must remap to ErrorLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.NotFound,
			level: logrus.InfoLevel,
			msg:   "NotFound must remap to InfoLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.FailedPrecondition,
			level: logrus.WarnLevel,
			msg:   "FailedPrecondition must remap to WarnLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.Unauthenticated,
			level: logrus.ErrorLevel,
			msg:   "Unauthenticated is overwritten to ErrorLevel with customCodeToLevel override, which probably didn't work",
		},
	} {
		s.buffer.Reset()
		_, err := s.Client.PingError(
			s.SimpleCtx(),
			&pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})
		assert.Error(s.T(), err, "each call here must return an error")
		msgs := s.getOutputJSONs()
		require.Len(s.T(), msgs, 1, "only the interceptor log message is printed in PingErr")
		m := msgs[0]
		assert.Contains(s.T(), m, `"grpc_service": "mwitkow.testproto.TestService"`, "all lines must contain service name")
		assert.Contains(s.T(), m, `"grpc_method": "PingError"`, "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"grpc_code": "%s"`, tcase.code.String()), "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"level": "%s"`, tcase.level.String()), tcase.msg)
	}
}

func (s *LogrusLoggingSuite) TestPingList_WithCustomTags() {
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
	assert.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		s.T()
		assert.Contains(s.T(), m, `"grpc_service": "mwitkow.testproto.TestService"`, "all lines must contain service name")
		assert.Contains(s.T(), m, `"grpc_method": "PingList"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"custom_string": "something"`, "all lines must contain `custom_string` set by AddFields")
		assert.Contains(s.T(), m, `"custom_int": 1337`, "all lines must contain `custom_int` set by AddFields")
		// request field extraction
		assert.Contains(s.T(), m, `"request.value": "something"`, "all lines must contain fields extracted from goodPing because of test.manual_extractfields.pb")
	}
	assert.Contains(s.T(), msgs[0], `"msg": "some pinglist"`, "handler's message must contain user message")
	assert.Contains(s.T(), msgs[1], `"msg": "finished streaming call"`, "interceptor message must contain string")
	assert.Contains(s.T(), msgs[1], `"level": "info"`, "OK error codes must be logged on info level.")
	assert.Contains(s.T(), msgs[1], `"grpc_time_ns":`, "interceptor log statement should contain execution time")
}

func (s *LogrusLoggingSuite) TestPingEmpty_WithMetadataTags() {
	_, err := s.Client.PingEmpty(s.SimpleCtx(), &pb_testproto.Empty{})
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "only the interceptor log message is printed in PingEmpty")
	m := msgs[0]
	assert.Contains(s.T(), m, `"middleware_1": 1410`, "the handler must contain fields from grpc_logging.Metadata calls")
	assert.Contains(s.T(), m, `"middleware_2": "some_content"`, "the handler must contain fields from grpc_logging.Metadata calls")
}
