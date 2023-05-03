// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type testDisposableFields map[string]string

func (f testDisposableFields) AssertField(t *testing.T, key, value string) testDisposableFields {
	t.Helper()

	require.Truef(t, len(f) > 0, "expected field %s = %v, but fields ended", key, value)
	assert.Equalf(t, value, f[key], "expected %s for %s", value, key)
	delete(f, key)
	return f
}

func (f testDisposableFields) AssertFieldNotEmpty(t *testing.T, key string) testDisposableFields {
	t.Helper()

	require.Truef(t, len(f) > 0, "expected field %s and some non-empty value, but fields ended", key)
	assert.Truef(t, f[key] != "", "%s is empty", key)
	delete(f, key)
	return f
}

func (f testDisposableFields) AssertNoMoreTags(t *testing.T) {
	t.Helper()

	assert.Lenf(t, f, 0, "expected no more fields in testDisposableFields but still got %v", f)
}

type LogLine struct {
	msg    string
	fields testDisposableFields
	lvl    logging.Level
}

type LogLines []LogLine

func (l LogLines) Len() int {
	return len(l)
}

func (l LogLines) Less(i, j int) bool {
	if l[i].fields[logging.ComponentFieldKey] != l[j].fields[logging.ComponentFieldKey] {
		return l[i].fields[logging.ComponentFieldKey] < l[j].fields[logging.ComponentFieldKey]
	}
	if l[i].msg != l[j].msg {
		return l[i].msg < l[j].msg
	}

	// We want to sort by counter which in string, so we need to parse it.
	a := testpb.PingResponse{}
	b := testpb.PingResponse{}
	_ = json.Unmarshal([]byte(l[i].fields["grpc.response.content"]), &a)
	_ = json.Unmarshal([]byte(l[j].fields["grpc.response.content"]), &b)
	if a.Counter != b.Counter {
		return a.Counter < b.Counter
	}

	_ = json.Unmarshal([]byte(l[i].fields["grpc.request.content"]), &a)
	_ = json.Unmarshal([]byte(l[j].fields["grpc.request.content"]), &b)
	return a.Counter < b.Counter
}

func (l LogLines) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type output struct {
	m     sync.Mutex
	lines LogLines
}

func (o *output) Lines() LogLines {
	o.m.Lock()
	defer o.m.Unlock()

	retLines := make(LogLines, len(o.lines))
	copy(retLines, o.lines)

	return retLines
}

func (o *output) Append(lines ...LogLine) {
	o.m.Lock()
	defer o.m.Unlock()

	o.lines = append(o.lines, lines...)
}
func (o *output) Reset() {
	o.m.Lock()
	defer o.m.Unlock()

	o.lines = o.lines[:0]
}

type mockLogger struct {
	o *output
}

func newMockLogger() *mockLogger {
	return &mockLogger{o: &output{}}
}

func (l *mockLogger) Log(_ context.Context, lvl logging.Level, msg string, args ...any) {
	line := LogLine{
		lvl:    lvl,
		msg:    msg,
		fields: map[string]string{},
	}

	// For test purposes we assume all args are string (or proto messages) and there is always pair.
	for i := 0; i < len(args); i += 2 {
		if v, ok := args[i+1].(string); ok {
			line.fields[args[i].(string)] = v
			continue
		}
		payload, err := protojson.Marshal(args[i+1].(proto.Message))
		if err != nil {
			panic(err)
		}
		// Trim spaces for deterministic output (https://github.com/goolang/protobuf/issues/1269).
		line.fields[args[i].(string)] = string(bytes.Replace(payload, []byte{' '}, []byte{}, -1))
	}
	l.o.Append(line)
}

type baseLoggingSuite struct {
	*testpb.InterceptorTestSuite
	logger *mockLogger
}

func (s *baseLoggingSuite) SetupTest() {
	s.logger.o.Reset()
}

func customClientCodeToLevel(c codes.Code) logging.Level {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return logging.LevelError
	}
	return logging.DefaultClientCodeToLevel(c)
}

type loggingClientServerSuite struct {
	*baseLoggingSuite
}

func customFields(_ context.Context) logging.Fields {
	// Add custom fields, one new and one that should be ignored as it duplicates the standard field.
	return logging.Fields{"custom-field", "yolo", logging.ServiceFieldKey, "something different"}
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	s := &loggingClientServerSuite{
		&baseLoggingSuite{
			logger: newMockLogger(),
			InterceptorTestSuite: &testpb.InterceptorTestSuite{
				TestService: &testpb.TestPingService{},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel), logging.WithFieldsFromContext(customFields)),
		),
		grpc.WithChainStreamInterceptor(
			logging.StreamClientInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel), logging.WithFieldsFromContext(customFields)),
		),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel), logging.WithFieldsFromContext(customFields))),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel), logging.WithFieldsFromContext(customFields))),
	}
	suite.Run(t, s)
}

func assertStandardFields(t *testing.T, kind string, f testDisposableFields, method string, typ interceptors.GRPCType) testDisposableFields {
	return f.AssertField(t, logging.SystemTag[0], logging.SystemTag[1]).
		AssertField(t, logging.ComponentFieldKey, kind).
		AssertField(t, logging.ServiceFieldKey, testpb.TestServiceFullName).
		AssertField(t, logging.MethodFieldKey, method).
		AssertField(t, logging.MethodTypeFieldKey, string(typ))
}

func (s *loggingClientServerSuite) TestPing() {
	ctx := logging.InjectFields(s.SimpleCtx(), logging.Fields{"grpc.request.value", testpb.GoodPing.Value})
	_, err := s.Client.Ping(ctx, testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	clientStartCallLogLine := lines[1]
	assert.Equal(s.T(), logging.LevelDebug, clientStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartCallLogLine.fields, "Ping", interceptors.Unary)

	serverStartCallLogLine := lines[3]
	assert.Equal(s.T(), logging.LevelDebug, serverStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartCallLogLine.fields, "Ping", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.LevelDebug, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "Ping", interceptors.Unary)
	serverFinishCallFields.AssertFieldNotEmpty(s.T(), "peer.address").
		AssertField(s.T(), "custom-field", "yolo").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.LevelDebug, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "Ping", interceptors.Unary)
	clientFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
		AssertField(s.T(), "grpc.request.value", "something").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingClientServerSuite) TestPingList() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}
	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	serverStartCallLogLine := lines[3]
	assert.Equal(s.T(), logging.LevelDebug, serverStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartCallLogLine.fields, "PingList", interceptors.ServerStream)

	clientStartCallLogLine := lines[1]
	assert.Equal(s.T(), logging.LevelDebug, clientStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartCallLogLine.fields, "PingList", interceptors.ServerStream)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.LevelDebug, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	serverFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
		AssertFieldNotEmpty(s.T(), "peer.address").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.LevelDebug, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	clientFinishCallFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertField(s.T(), "custom-field", "yolo").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingClientServerSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  codes.Code
		level logging.Level
		msg   string
	}{
		{
			code:  codes.Internal,
			level: logging.LevelWarn,
			msg:   "Internal must remap to WarnLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.NotFound,
			level: logging.LevelDebug,
			msg:   "NotFound must remap to DebugLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.FailedPrecondition,
			level: logging.LevelDebug,
			msg:   "FailedPrecondition must remap to DebugLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.Unauthenticated,
			level: logging.LevelError,
			msg:   "Unauthenticated is overwritten to ErrorLevel with customClientCodeToLevel override, which probably didn't work",
		},
	} {
		s.SetupTest()
		s.T().Run(tcase.msg, func(t *testing.T) {
			_, err := s.Client.PingError(
				s.SimpleCtx(),
				&testpb.PingErrorRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})
			require.Error(t, err, "each call here must return an error")
			lines := s.logger.o.Lines()
			sort.Sort(lines)
			require.Len(t, lines, 4)

			serverFinishCallLogLine := lines[2]
			assert.Equal(t, tcase.level, serverFinishCallLogLine.lvl)
			assert.Equal(t, "finished call", serverFinishCallLogLine.msg)
			serverFinishCallFields := assertStandardFields(t, logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingError", interceptors.Unary)
			serverFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
				AssertFieldNotEmpty(t, "peer.address").
				AssertFieldNotEmpty(t, "grpc.start_time").
				AssertFieldNotEmpty(t, "grpc.request.deadline").
				AssertField(t, "grpc.code", tcase.code.String()).
				AssertField(t, "grpc.error", fmt.Sprintf("rpc error: code = %s desc = Userspace error.", tcase.code.String())).
				AssertFieldNotEmpty(t, "grpc.time_ms").AssertNoMoreTags(t)

			clientFinishCallLogLine := lines[0]
			assert.Equal(t, tcase.level, clientFinishCallLogLine.lvl)
			assert.Equal(t, "finished call", clientFinishCallLogLine.msg)
			clientFinishCallFields := assertStandardFields(t, logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingError", interceptors.Unary)
			clientFinishCallFields.AssertField(s.T(), "custom-field", "yolo").
				AssertFieldNotEmpty(t, "grpc.start_time").
				AssertFieldNotEmpty(t, "grpc.request.deadline").
				AssertField(t, "grpc.code", tcase.code.String()).
				AssertField(t, "grpc.error", fmt.Sprintf("rpc error: code = %s desc = Userspace error.", tcase.code.String())).
				AssertFieldNotEmpty(t, "grpc.time_ms").AssertNoMoreTags(t)
		})
	}
}

type loggingCustomDurationSuite struct {
	*baseLoggingSuite
}

func TestCustomDurationSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	s := &loggingCustomDurationSuite{
		baseLoggingSuite: &baseLoggingSuite{
			logger: newMockLogger(),
			InterceptorTestSuite: &testpb.InterceptorTestSuite{
				TestService: &testpb.TestPingService{},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
		grpc.WithStreamInterceptor(logging.StreamClientInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
	}
	suite.Run(t, s)
}

func (s *loggingCustomDurationSuite) TestPing_HasOverriddenDuration() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[3]
	assert.Equal(s.T(), logging.LevelInfo, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "Ping", interceptors.Unary)

	clientStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.LevelDebug, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "Ping", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.LevelInfo, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "Ping", interceptors.Unary)
	serverFinishCallFields.AssertFieldNotEmpty(s.T(), "peer.address").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.LevelDebug, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "Ping", interceptors.Unary)
	clientFinishCallFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())
}

func (s *loggingCustomDurationSuite) TestPingList_HasOverriddenDuration() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[3]
	assert.Equal(s.T(), logging.LevelInfo, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "PingList", interceptors.ServerStream)

	clientStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.LevelDebug, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "PingList", interceptors.ServerStream)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.LevelInfo, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	serverFinishCallFields.AssertFieldNotEmpty(s.T(), "peer.address").
		AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[0]
	assert.Equal(s.T(), logging.LevelDebug, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	clientFinishCallFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertField(s.T(), "grpc.code", "OK").
		AssertFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())
}

type loggingPayloadSuite struct {
	*baseLoggingSuite
}

func TestPayloadSuite(t *testing.T) {
	s := &loggingPayloadSuite{
		baseLoggingSuite: &baseLoggingSuite{
			logger: newMockLogger(),
			InterceptorTestSuite: &testpb.InterceptorTestSuite{
				TestService: &testpb.TestPingService{},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(
			s.logger,
			logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
			logging.WithLevels(logging.DefaultClientCodeToLevel),
		)),
		grpc.WithStreamInterceptor(logging.StreamClientInterceptor(s.logger,
			logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
			logging.WithLevels(logging.DefaultClientCodeToLevel),
		)),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc.StreamInterceptor(logging.StreamServerInterceptor(s.logger,
			logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
			logging.WithLevels(logging.DefaultServerCodeToLevel),
		)),
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(s.logger,
			logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
			logging.WithLevels(logging.DefaultServerCodeToLevel),
		)),
	}
	suite.Run(t, s)
}

func (s *loggingPayloadSuite) TestPing_LogsBothRequestAndResponse() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	lines := s.logger.o.Lines()
	require.Len(s.T(), lines, 4)
	s.assertPayloadLogLinesForMessage(lines, "Ping", interceptors.Unary)
}

func (s *loggingPayloadSuite) TestPingError_LogsOnlyRequestsOnError() {
	_, err := s.Client.PingError(s.SimpleCtx(), &testpb.PingErrorRequest{Value: "something", ErrorCodeReturned: uint32(4)})
	require.Error(s.T(), err, "there must be an error on an unsuccessful call")

	lines := s.logger.o.Lines()
	sort.Sort(lines)
	require.Len(s.T(), lines, 2) // Only one client and one server for request, since no response (error).
	clientRequestLogLine := lines[0]
	assert.Equal(s.T(), logging.DefaultClientCodeToLevel(codes.OK), clientRequestLogLine.lvl)
	assert.Equal(s.T(), "request sent", clientRequestLogLine.msg)
	clientRequestFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientRequestLogLine.fields, "PingError", interceptors.Unary)

	clientRequestFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
		AssertFieldNotEmpty(s.T(), "grpc.send.duration").
		AssertField(s.T(), "grpc.request.content", `{"value":"something","errorCodeReturned":4}`).
		AssertFieldNotEmpty(s.T(), "grpc.request.deadline").AssertNoMoreTags(s.T())
}

func (s *loggingPayloadSuite) TestPingStream_LogsAllRequestsAndResponses() {
	messagesExpected := 20
	stream, err := s.Client.PingStream(s.SimpleCtx())

	require.NoError(s.T(), err, "no error on stream creation")
	for i := 0; i < messagesExpected; i++ {
		require.NoError(s.T(), stream.Send(testpb.GoodPingStream), "sending must succeed")

		pong := &testpb.PingResponse{}
		err := stream.RecvMsg(pong)
		require.NoError(s.T(), err, "no error on receive")
	}
	require.NoError(s.T(), stream.CloseSend(), "no error on send stream")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	require.NoError(s.T(), waitUntil(200*time.Millisecond, ctx.Done(), func() error {
		got := len(s.logger.o.Lines())
		if got >= 4*messagesExpected {
			return nil
		}
		return errors.Errorf("not enough log lines, waiting; got: %v", got)
	}))
	s.assertPayloadLogLinesForMessage(s.logger.o.Lines(), "PingStream", interceptors.BidiStream)
}

func (s *loggingPayloadSuite) assertPayloadLogLinesForMessage(lines LogLines, method string, typ interceptors.GRPCType) {
	// Order matter for assertion, we don't rely on it though, so sort.
	sort.Sort(lines)

	repetitions := len(lines) / 4
	curr := 0
	for i := curr; i < repetitions; i++ {
		clientRequestLogLine := lines[i]
		assert.Equal(s.T(), logging.DefaultClientCodeToLevel(codes.OK), clientRequestLogLine.lvl)
		assert.Equal(s.T(), "request sent", clientRequestLogLine.msg)
		clientRequestFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientRequestLogLine.fields, method, typ)
		clientRequestFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.send.duration").
			AssertField(s.T(), "grpc.request.content", `{"value":"something","sleepTimeMs":9999}`).
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline").AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		clientResponseLogLine := lines[i]
		assert.Equal(s.T(), logging.DefaultClientCodeToLevel(codes.OK), clientResponseLogLine.lvl)
		assert.Equal(s.T(), "response received", clientResponseLogLine.msg)
		clientResponseFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientResponseLogLine.fields, method, typ)
		clientResponseFields = clientResponseFields.AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.recv.duration").
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline")
		if i-curr == 0 {
			clientResponseFields = clientResponseFields.AssertField(s.T(), "grpc.response.content", `{"value":"something"}`)
		} else {
			clientResponseFields = clientResponseFields.AssertField(s.T(), "grpc.response.content", fmt.Sprintf(`{"value":"something","counter":%v}`, i-curr))
		}
		clientResponseFields.AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		serverRequestLogLine := lines[i]
		assert.Equal(s.T(), logging.DefaultServerCodeToLevel(codes.OK), serverRequestLogLine.lvl)
		assert.Equal(s.T(), "request received", serverRequestLogLine.msg)
		serverRequestFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverRequestLogLine.fields, method, typ)
		serverRequestFields.AssertFieldNotEmpty(s.T(), "peer.address").
			AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.recv.duration").
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline").
			AssertField(s.T(), "grpc.request.content", `{"value":"something","sleepTimeMs":9999}`).AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		serverResponseLogLine := lines[i]
		assert.Equal(s.T(), logging.DefaultServerCodeToLevel(codes.OK), serverResponseLogLine.lvl)
		assert.Equal(s.T(), "response sent", serverResponseLogLine.msg)
		serverResponseFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverResponseLogLine.fields, method, typ)
		serverResponseFields = serverResponseFields.AssertFieldNotEmpty(s.T(), "peer.address").
			AssertFieldNotEmpty(s.T(), "grpc.start_time").
			AssertFieldNotEmpty(s.T(), "grpc.send.duration").
			AssertFieldNotEmpty(s.T(), "grpc.request.deadline")
		if i-curr == 0 {
			serverResponseFields = serverResponseFields.AssertField(s.T(), "grpc.response.content", `{"value":"something"}`)
		} else {
			serverResponseFields = serverResponseFields.AssertField(s.T(), "grpc.response.content", fmt.Sprintf(`{"value":"something","counter":%v}`, i-curr))
		}
		serverResponseFields.AssertNoMoreTags(s.T())
	}
}

// waitUntil executes f every interval seconds until timeout or no error is returned from f.
func waitUntil(interval time.Duration, stopc <-chan struct{}, f func() error) error {
	tick := time.NewTicker(interval)
	defer tick.Stop()

	var err error
	for {
		if err = f(); err == nil {
			return nil
		}
		select {
		case <-stopc:
			return err
		case <-tick.C:
		}
	}
}
