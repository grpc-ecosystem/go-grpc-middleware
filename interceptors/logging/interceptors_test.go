package logging_test

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

var (
	goodPing = &testpb.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type testDisposableFields map[string]string

func (f testDisposableFields) AssertNextField(t *testing.T, key, value string) testDisposableFields {
	require.Truef(t, len(f) > 0, "expected field %s = %v, but fields ended", key, value)
	assert.Equalf(t, value, f[key], "expected %s for %s", value, key)
	delete(f, key)
	return f
}

func (f testDisposableFields) AssertNextFieldNotEmpty(t *testing.T, key string) testDisposableFields {
	require.Truef(t, len(f) > 0, "expected field %s and some non-empty value, but fields ended", key)
	assert.Truef(t, f[key] != "", "%s is empty", key)
	delete(f, key)
	return f
}

func (f testDisposableFields) AssertNoMoreTags(t *testing.T) {
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
	//_ , aok = l[i].fields["grpc.response.content"]
	//_ ,baok = l[i].fields["grpc.response.content"]
	//if

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

type baseMockLogger struct {
	// Shared. It's slice on purpose to find potential duplicates.
	lines []LogLine
	m     sync.Mutex
}

func (l *baseMockLogger) Add(line LogLine) {
	l.m.Lock()
	defer l.m.Unlock()

	l.lines = append(l.lines, line)
}

func (l *baseMockLogger) Lines() []LogLine {
	l.m.Lock()
	defer l.m.Unlock()

	return l.lines
}

type mockLogger struct {
	*baseMockLogger

	fields logging.Fields
}

func (l *mockLogger) Log(lvl logging.Level, msg string) {
	line := LogLine{
		lvl:    lvl,
		msg:    msg,
		fields: map[string]string{},
	}
	for i := 0; i < len(l.fields); i += 2 {
		line.fields[l.fields[i]] = l.fields[i+1]
	}
	l.Add(line)
}

func (l *mockLogger) With(fields ...string) logging.Logger {
	// Append twice to copy slice, so we don't reuse array.
	return &mockLogger{baseMockLogger: l.baseMockLogger, fields: append(append(logging.Fields{}, l.fields...), fields...)}
}

type baseLoggingSuite struct {
	*grpctesting.InterceptorTestSuite
	logger *mockLogger
}

func (s *baseLoggingSuite) SetupTest() {
	s.logger.fields = s.logger.fields[:0]
	s.logger.lines = s.logger.lines[:0]
}

func customClientCodeToLevel(c codes.Code) logging.Level {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return logging.ERROR
	}
	return logging.DefaultClientCodeToLevel(c)
}

type loggingClientServerSuite struct {
	*baseLoggingSuite
}

func TestSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	s := &loggingClientServerSuite{
		&baseLoggingSuite{
			logger: &mockLogger{baseMockLogger: &baseMockLogger{}},
			InterceptorTestSuite: &grpctesting.InterceptorTestSuite{
				TestService: &grpctesting.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel))),
		grpc.WithStreamInterceptor(logging.StreamClientInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel))),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.StreamServerInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel))),
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.UnaryServerInterceptor(s.logger, logging.WithLevels(customClientCodeToLevel))),
	}
	suite.Run(t, s)
}

func assertStandardFields(t *testing.T, kind string, f testDisposableFields, method string, typ interceptors.GRPCType) testDisposableFields {
	return f.AssertNextField(t, logging.SystemTag[0], logging.SystemTag[1]).
		AssertNextField(t, logging.ComponentFieldKey, kind).
		AssertNextField(t, logging.ServiceFieldKey, "grpc_middleware.testpb.TestService").
		AssertNextField(t, logging.MethodFieldKey, method).
		AssertNextField(t, logging.MethodTypeFieldKey, string(typ))
}

func (s *loggingClientServerSuite) TestPing() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	lines := s.logger.Lines()
	require.Len(s.T(), lines, 4)

	clientStartCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartCallLogLine.fields, "Ping", interceptors.Unary)

	serverStartCallLogLine := lines[1]
	assert.Equal(s.T(), logging.DEBUG, serverStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartCallLogLine.fields, "Ping", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.DEBUG, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "Ping", interceptors.Unary)
	serverFinishCallFields.AssertNextField(s.T(), "grpc.request.value", "something").
		AssertNextFieldNotEmpty(s.T(), "peer.address").
		AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[3]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "Ping", interceptors.Unary)
	clientFinishCallFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingClientServerSuite) TestPingList() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}
	lines := s.logger.Lines()
	require.Len(s.T(), lines, 4)

	serverStartCallLogLine := lines[1]
	assert.Equal(s.T(), logging.DEBUG, serverStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartCallLogLine.fields, "PingList", interceptors.ServerStream)

	clientStartCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientStartCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartCallLogLine.fields, "PingList", interceptors.ServerStream)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.DEBUG, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	serverFinishCallFields.AssertNextField(s.T(), "grpc.request.value", "something").
		AssertNextFieldNotEmpty(s.T(), "peer.address").
		AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[3]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	clientFinishCallFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingClientServerSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  codes.Code
		level logging.Level
		msg   string
	}{
		{
			code:  codes.Internal,
			level: logging.WARNING,
			msg:   "Internal must remap to WarnLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.NotFound,
			level: logging.DEBUG,
			msg:   "NotFound must remap to DebugLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.FailedPrecondition,
			level: logging.DEBUG,
			msg:   "FailedPrecondition must remap to DebugLevel in DefaultClientCodeToLevel",
		},
		{
			code:  codes.Unauthenticated,
			level: logging.ERROR,
			msg:   "Unauthenticated is overwritten to ErrorLevel with customClientCodeToLevel override, which probably didn't work",
		},
	} {
		s.SetupTest()
		s.T().Run(tcase.msg, func(t *testing.T) {
			_, err := s.Client.PingError(
				s.SimpleCtx(),
				&testpb.PingRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})
			require.Error(t, err, "each call here must return an error")
			lines := s.logger.Lines()
			require.Len(t, lines, 4)

			serverFinishCallLogLine := lines[2]
			assert.Equal(t, tcase.level, serverFinishCallLogLine.lvl)
			assert.Equal(t, "finished call", serverFinishCallLogLine.msg)
			serverFinishCallFields := assertStandardFields(t, logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingError", interceptors.Unary)
			serverFinishCallFields.AssertNextField(t, "grpc.request.value", "something").
				AssertNextFieldNotEmpty(t, "peer.address").
				AssertNextFieldNotEmpty(t, "grpc.start_time").
				AssertNextFieldNotEmpty(t, "grpc.request.deadline").
				AssertNextField(t, "grpc.code", tcase.code.String()).
				AssertNextField(t, "grpc.error", fmt.Sprintf("rpc error: code = %s desc = Userspace error.", tcase.code.String())).
				AssertNextFieldNotEmpty(t, "grpc.time_ms").AssertNoMoreTags(t)

			clientFinishCallLogLine := lines[3]
			assert.Equal(t, tcase.level, clientFinishCallLogLine.lvl)
			assert.Equal(t, "finished call", clientFinishCallLogLine.msg)
			clientFinishCallFields := assertStandardFields(t, logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingError", interceptors.Unary)
			clientFinishCallFields.AssertNextFieldNotEmpty(t, "grpc.start_time").
				AssertNextFieldNotEmpty(t, "grpc.request.deadline").
				AssertNextField(t, "grpc.code", tcase.code.String()).
				AssertNextField(t, "grpc.error", fmt.Sprintf("rpc error: code = %s desc = Userspace error.", tcase.code.String())).
				AssertNextFieldNotEmpty(t, "grpc.time_ms").AssertNoMoreTags(t)
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
			logger: &mockLogger{baseMockLogger: &baseMockLogger{}},
			InterceptorTestSuite: &grpctesting.InterceptorTestSuite{
				TestService: &grpctesting.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
		grpc.WithStreamInterceptor(logging.StreamClientInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.StreamServerInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.UnaryServerInterceptor(s.logger, logging.WithDurationField(logging.DurationToDurationField))),
	}
	suite.Run(t, s)
}

func (s *loggingCustomDurationSuite) TestPing_HasOverriddenDuration() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")

	lines := s.logger.Lines()
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.INFO, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "Ping", interceptors.Unary)

	clientStartedCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "Ping", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.INFO, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "Ping", interceptors.Unary)
	serverFinishCallFields.AssertNextField(s.T(), "grpc.request.value", "something").
		AssertNextFieldNotEmpty(s.T(), "peer.address").
		AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[3]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "Ping", interceptors.Unary)
	clientFinishCallFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())
}

func (s *loggingCustomDurationSuite) TestPingList_HasOverriddenDuration() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}

	lines := s.logger.Lines()
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.INFO, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "PingList", interceptors.ServerStream)

	clientStartedCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "PingList", interceptors.ServerStream)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.INFO, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	serverFinishCallFields.AssertNextField(s.T(), "grpc.request.value", "something").
		AssertNextFieldNotEmpty(s.T(), "peer.address").
		AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[3]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingList", interceptors.ServerStream)
	clientFinishCallFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "OK").
		AssertNextFieldNotEmpty(s.T(), "grpc.duration").AssertNoMoreTags(s.T())
}

type loggingCustomDeciderSuite struct {
	*baseLoggingSuite
}

func TestCustomDeciderSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skip("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	opts := logging.WithDecider(func(method string) bool {
		return method == "/grpc_middleware.testpb.TestService/PingError"
	})

	s := &loggingCustomDeciderSuite{
		baseLoggingSuite: &baseLoggingSuite{
			logger: &mockLogger{baseMockLogger: &baseMockLogger{}},
			InterceptorTestSuite: &grpctesting.InterceptorTestSuite{
				TestService: &grpctesting.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(s.logger, opts)),
		grpc.WithStreamInterceptor(logging.StreamClientInterceptor(s.logger, opts)),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.StreamServerInterceptor(s.logger, opts)),
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.UnaryServerInterceptor(s.logger, opts)),
	}
	suite.Run(t, s)
}

func (s *loggingCustomDeciderSuite) TestPing_HasCustomDecider() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	require.Len(s.T(), s.logger.Lines(), 0) // Decider should suppress.
}

func (s *loggingCustomDeciderSuite) TestPingError_HasCustomDecider() {
	code := codes.NotFound

	_, err := s.Client.PingError(
		s.SimpleCtx(),
		&testpb.PingRequest{Value: "something", ErrorCodeReturned: uint32(code)})
	require.Error(s.T(), err, "each call here must return an error")

	lines := s.logger.Lines()
	require.Len(s.T(), lines, 4)

	serverStartedCallLogLine := lines[1]
	assert.Equal(s.T(), logging.INFO, serverStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", serverStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindServerFieldValue, serverStartedCallLogLine.fields, "PingError", interceptors.Unary)

	clientStartedCallLogLine := lines[0]
	assert.Equal(s.T(), logging.DEBUG, clientStartedCallLogLine.lvl)
	assert.Equal(s.T(), "started call", clientStartedCallLogLine.msg)
	_ = assertStandardFields(s.T(), logging.KindClientFieldValue, clientStartedCallLogLine.fields, "PingError", interceptors.Unary)

	serverFinishCallLogLine := lines[2]
	assert.Equal(s.T(), logging.INFO, serverFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", serverFinishCallLogLine.msg)
	serverFinishCallFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverFinishCallLogLine.fields, "PingError", interceptors.Unary)
	serverFinishCallFields.AssertNextField(s.T(), "grpc.request.value", "something").
		AssertNextFieldNotEmpty(s.T(), "peer.address").
		AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "NotFound").
		AssertNextField(s.T(), "grpc.error", "rpc error: code = NotFound desc = Userspace error.").
		AssertNextFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())

	clientFinishCallLogLine := lines[3]
	assert.Equal(s.T(), logging.DEBUG, clientFinishCallLogLine.lvl)
	assert.Equal(s.T(), "finished call", clientFinishCallLogLine.msg)
	clientFinishCallFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientFinishCallLogLine.fields, "PingError", interceptors.Unary)
	clientFinishCallFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
		AssertNextField(s.T(), "grpc.code", "NotFound").
		AssertNextField(s.T(), "grpc.error", "rpc error: code = NotFound desc = Userspace error.").
		AssertNextFieldNotEmpty(s.T(), "grpc.time_ms").AssertNoMoreTags(s.T())
}

func (s *loggingCustomDeciderSuite) TestPingList_HasCustomDecider() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}

	require.Len(s.T(), s.logger.Lines(), 0) // Decider should suppress.
}
