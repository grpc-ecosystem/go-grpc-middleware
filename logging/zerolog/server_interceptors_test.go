package zerolog_test

import (
	"io"
	"testing"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zerolog "github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestZerologServerSuite(t *testing.T) {
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithLevels(customCodeToLevel),
	}
	b := newZerologBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zerolog.StreamServerInterceptor(b.logger, opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zerolog.UnaryServerInterceptor(b.logger, opts...)),
	}
	suite.Run(t, &zerologServerSuite{b})
}

type zerologServerSuite struct {
	*zerologBaseSuite
}

func (s *zerologServerSuite) TestPing_WithCustomTags() {
	deadline := time.Now().Add(3 * time.Second)
	_, err := s.Client.Ping(s.DeadlineCtx(deadline), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "two log statements should be logged")

	for _, m := range msgs {
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
		assert.Equal(s.T(), m["grpc.method"], "Ping", "all lines must contain the correct method name")
		assert.Equal(s.T(), m["span.kind"], "server", "all lines must contain the kind of call (server)")
		assert.Equal(s.T(), m["custom_tags.string"], "something", "all lines must contain `custom_tags.string` with expected value")
		assert.Equal(s.T(), m["grpc.request.value"], "something", "all lines must contain the correct request value")
		assert.Equal(s.T(), m["custom_field"], "custom_value", "all lines must contain `custom_field` with the correct value")

		assert.Contains(s.T(), m, "custom_tags.int", "all lines must contain `custom_tags.int`")
		require.Contains(s.T(), m, "grpc.start_time", "all lines must contain the start time of the call")
		_, err := time.Parse(time.RFC3339, m["grpc.start_time"].(string))
		assert.NoError(s.T(), err, "should be able to parse start time as RFC3339")

		require.Contains(s.T(), m, "grpc.request.deadline", "all lines must contain the deadline of the call")
		_, err = time.Parse(time.RFC3339, m["grpc.request.deadline"].(string))
		require.NoError(s.T(), err, "should be able to parse deadline as RFC3339")
		assert.Equal(s.T(), m["grpc.request.deadline"], deadline.Format(time.RFC3339), "should have the same deadline that was set by the caller")
	}

	assert.Equal(s.T(), msgs[0]["message"], "some ping", "first message must contain the correct user message")
	assert.Equal(s.T(), msgs[1]["message"], "finished unary call with code OK", "second message must contain the correct user message")
	assert.Equal(s.T(), msgs[1]["level"], "info", "OK codes must be logged on info level.")

	assert.Contains(s.T(), msgs[1], "grpc.time_ms", "interceptor log statement should contain execution time")
}

func (s *zerologServerSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  codes.Code
		level zerolog.Level
		msg   string
	}{
		{
			code:  codes.Internal,
			level: zerolog.ErrorLevel,
			msg:   "Internal must remap to ErrorLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.NotFound,
			level: zerolog.InfoLevel,
			msg:   "NotFound must remap to InfoLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.FailedPrecondition,
			level: zerolog.WarnLevel,
			msg:   "FailedPrecondition must remap to WarnLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.Unauthenticated,
			level: zerolog.ErrorLevel,
			msg:   "Unauthenticated is overwritten to ErrorLevel with customCodeToLevel override, which probably didn't work",
		},
	} {
		s.buffer.Reset()
		_, err := s.Client.PingError(
			s.SimpleCtx(),
			&pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})
		require.Error(s.T(), err, "each call here must return an error")

		msgs := s.getOutputJSONs()
		require.Len(s.T(), msgs, 1, "only the interceptor log message is printed in PingErr")
		m := msgs[0]
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
		assert.Equal(s.T(), m["grpc.method"], "PingError", "all lines must contain the correct method name")
		assert.Equal(s.T(), m["grpc.code"], tcase.code.String(), "a gRPC code must be present")
		assert.Equal(s.T(), m["level"], tcase.level.String(), tcase.msg)
		assert.Equal(s.T(), m["message"], "finished unary call with code "+tcase.code.String(), "must have the correct finish message")

		require.Contains(s.T(), m, "grpc.start_time", "all lines must contain a start time for the call")
		_, err = time.Parse(time.RFC3339, m["grpc.start_time"].(string))
		assert.NoError(s.T(), err, "should be able to parse the start time as RFC3339")

		require.Contains(s.T(), m, "grpc.request.deadline", "all lines must contain the deadline of the call")
		_, err = time.Parse(time.RFC3339, m["grpc.request.deadline"].(string))
		require.NoError(s.T(), err, "should be able to parse deadline as RFC3339")
	}
}

func (s *zerologServerSuite) TestPingList_WithCustomTags() {
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
	require.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
		assert.Equal(s.T(), m["grpc.method"], "PingList", "all lines must contain the correct method name")
		assert.Equal(s.T(), m["span.kind"], "server", "all lines must contain the kind of call (server)")
		assert.Equal(s.T(), m["custom_tags.string"], "something", "all lines must contain the correct `custom_tags.string`")
		assert.Equal(s.T(), m["grpc.request.value"], "something", "all lines must contain the correct request value")

		assert.Contains(s.T(), m, "custom_tags.int", "all lines must contain `custom_tags.int`")
		require.Contains(s.T(), m, "grpc.start_time", "all lines must contain the start time for the call")
		_, err := time.Parse(time.RFC3339, m["grpc.start_time"].(string))
		assert.NoError(s.T(), err, "should be able to parse start time as RFC3339")

		require.Contains(s.T(), m, "grpc.request.deadline", "all lines must contain the deadline of the call")
		_, err = time.Parse(time.RFC3339, m["grpc.request.deadline"].(string))
		require.NoError(s.T(), err, "should be able to parse deadline as RFC3339")
	}

	assert.Equal(s.T(), msgs[0]["message"], "some pinglist", "msg must be the correct message")
	assert.Equal(s.T(), msgs[1]["message"], "finished streaming call with code OK", "msg must be the correct message")
	assert.Equal(s.T(), msgs[1]["level"], "info", "OK codes must be logged on info level.")

	assert.Contains(s.T(), msgs[1], "grpc.time_ms", "interceptor log statement should contain execution time")
}

func TestZerologServerOverrideSuite(t *testing.T) {
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithDurationField(grpc_zerolog.DurationToDurationField),
	}
	b := newZerologBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zerolog.StreamServerInterceptor(b.logger, opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zerolog.UnaryServerInterceptor(b.logger, opts...)),
	}
	suite.Run(t, &zerologServerOverrideSuite{b})
}

type zerologServerOverrideSuite struct {
	*zerologBaseSuite
}

func (s *zerologServerOverrideSuite) TestPing_HasOverriddenDuration() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain service name")
		assert.Equal(s.T(), m["grpc.method"], "Ping", "all lines must contain method name")
	}

	assert.Equal(s.T(), msgs[0]["message"], "some ping", "first message must be correct")
	assert.NotContains(s.T(), msgs[0], "grpc.time_ms", "first message must not contain default duration")
	assert.NotContains(s.T(), msgs[0], "grpc.duration", "first message must not contain overridden duration")

	assert.Equal(s.T(), msgs[1]["message"], "finished unary call with code OK", "second message must be correct")
	assert.Equal(s.T(), msgs[1]["level"], "info", "second must be logged on info level.")
	assert.NotContains(s.T(), msgs[1], "grpc.time_ms", "second message must not contain default duration")
	assert.Contains(s.T(), msgs[1], "grpc.duration", "second message must contain overridden duration")
}

func (s *zerologServerOverrideSuite) TestPingList_HasOverriddenDuration() {
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
	require.Len(s.T(), msgs, 2, "two log statements should be logged")

	for _, m := range msgs {
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain service name")
		assert.Equal(s.T(), m["grpc.method"], "PingList", "all lines must contain method name")
	}

	assert.Equal(s.T(), msgs[0]["message"], "some pinglist", "first message must contain user message")
	assert.NotContains(s.T(), msgs[0], "grpc.time_ms", "first message must not contain default duration")
	assert.NotContains(s.T(), msgs[0], "grpc.duration", "first message must not contain overridden duration")

	assert.Equal(s.T(), msgs[1]["message"], "finished streaming call with code OK", "second message must contain correct message")
	assert.Equal(s.T(), msgs[1]["level"], "info", "second message must be logged on info level.")
	assert.NotContains(s.T(), msgs[1], "grpc.time_ms", "second message must not contain default duration")
	assert.Contains(s.T(), msgs[1], "grpc.duration", "second message must contain overridden duration")
}

func TestZerologServerOverrideDeciderSuite(t *testing.T) {
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithDecider(func(method string, err error) bool {
			if err != nil && method == "/mwitkow.testproto.TestService/PingError" {
				return true
			}

			return false
		}),
	}
	b := newZerologBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zerolog.StreamServerInterceptor(b.logger, opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zerolog.UnaryServerInterceptor(b.logger, opts...)),
	}
	suite.Run(t, &zerologServerOverrideDeciderSuite{b})
}

type zerologServerOverrideDeciderSuite struct {
	*zerologBaseSuite
}

func (s *zerologServerOverrideDeciderSuite) TestPing_HasOverriddenDecider() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "single log statements should be logged")
	assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain service name")
	assert.Equal(s.T(), msgs[0]["grpc.method"], "Ping", "all lines must contain method name")
	assert.Equal(s.T(), msgs[0]["message"], "some ping", "handler's message must contain user message")
}

func (s *zerologServerOverrideDeciderSuite) TestPingError_HasOverriddenDecider() {
	code := codes.NotFound
	level := zerolog.InfoLevel
	msg := "NotFound must remap to InfoLevel in DefaultCodeToLevel"

	s.buffer.Reset()
	_, err := s.Client.PingError(
		s.SimpleCtx(),
		&pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(code)})
	require.Error(s.T(), err, "each call here must return an error")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "only the interceptor log message is printed in PingErr")
	m := msgs[0]
	assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain service name")
	assert.Equal(s.T(), m["grpc.method"], "PingError", "all lines must contain method name")
	assert.Equal(s.T(), m["grpc.code"], code.String(), "all lines must correct gRPC code")
	assert.Equal(s.T(), m["level"], level.String(), msg)
}

func (s *zerologServerOverrideDeciderSuite) TestPingList_HasOverriddenDecider() {
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
	require.Len(s.T(), msgs, 1, "single log statements should be logged")
	assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain service name")
	assert.Equal(s.T(), msgs[0]["grpc.method"], "PingList", "all lines must contain method name")
	assert.Equal(s.T(), msgs[0]["message"], "some pinglist", "handler's message must contain user message")

	assert.NotContains(s.T(), msgs[0], "grpc.time_ms", "handler's message must not contain default duration")
	assert.NotContains(s.T(), msgs[0], "grpc.duration", "handler's message must not contain overridden duration")
}

func TestZerologServerMessageProducerSuite(t *testing.T) {
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithMessageProducer(StubMessageProducer),
	}
	b := newZerologBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zerolog.StreamServerInterceptor(b.logger, opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zerolog.UnaryServerInterceptor(b.logger, opts...)),
	}
	suite.Run(t, &zerologServerMessageProducerSuite{b})
}

type zerologServerMessageProducerSuite struct {
	*zerologBaseSuite
}

func (s *zerologServerMessageProducerSuite) TestPing_HasMessageProducer() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "single log statements should be logged")
	assert.Equal(s.T(), msgs[0]["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain service name")
	assert.Equal(s.T(), msgs[0]["grpc.method"], "Ping", "all lines must contain method name")
	assert.Equal(s.T(), msgs[1]["message"], "custom message", "user defined message producer must be used")

	assert.Equal(s.T(), msgs[0]["message"], "some ping", "handler's message must contain user message")
}
