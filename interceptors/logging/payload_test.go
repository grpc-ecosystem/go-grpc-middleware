package logging_test

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

type loggingPayloadSuite struct {
	*baseLoggingSuite
}

func TestPayloadSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }

	s := &loggingPayloadSuite{
		baseLoggingSuite: &baseLoggingSuite{
			logger: &mockLogger{sharedResults: &sharedResults{}},
			InterceptorTestSuite: &grpctesting.InterceptorTestSuite{
				TestService: &grpctesting.TestPingService{T: t},
			},
		},
	}
	s.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(logging.PayloadUnaryClientInterceptor(s.logger, alwaysLoggingDeciderClient)),
		grpc.WithStreamInterceptor(logging.PayloadStreamClientInterceptor(s.logger, alwaysLoggingDeciderClient)),
	}
	s.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.PayloadStreamServerInterceptor(s.logger, alwaysLoggingDeciderServer)),
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.PayloadUnaryServerInterceptor(s.logger, alwaysLoggingDeciderServer)),
	}
	suite.Run(t, s)
}

func (s *loggingPayloadSuite) TestPing_LogsBothRequestAndResponse() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an error on a successful call")

	lines := s.logger.Lines()
	require.Len(s.T(), lines, 4)
	s.assertPayloadLogLinesForMessage(lines, "Ping", interceptors.Unary)
}

func (s *loggingPayloadSuite) assertPayloadLogLinesForMessage(lines LogLines, method string, typ interceptors.GRPCType) {
	// Order matter for assertion, we don't rely on it though, so sort.
	sort.Sort(lines)

	repetitions := len(lines) / 4
	curr := 0
	for i := curr; i < repetitions; i++ {
		clientRequestLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, clientRequestLogLine.lvl)
		assert.Equal(s.T(), "request payload logged as grpc.request.content field", clientRequestLogLine.msg)
		clientRequestFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientRequestLogLine.fields, method, typ)
		clientRequestFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
			AssertNextFieldNotEmpty(s.T(), "grpc.send.duration").
			AssertNextField(s.T(), "grpc.request.content", `{"value":"something","sleepTimeMs":9999}`).
			AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		clientResponseLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, clientResponseLogLine.lvl)
		assert.Equal(s.T(), "response payload logged as grpc.response.content field", clientResponseLogLine.msg)
		clientResponseFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientResponseLogLine.fields, method, typ)
		clientResponseFields = clientResponseFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
			AssertNextFieldNotEmpty(s.T(), "grpc.recv.duration").
			AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline")
		if i-curr == 0 {
			clientResponseFields = clientResponseFields.AssertNextField(s.T(), "grpc.response.content", `{"Value":"something"}`)
		} else {
			clientResponseFields = clientResponseFields.AssertNextField(s.T(), "grpc.response.content", fmt.Sprintf(`{"Value":"something","counter":%v}`, i-curr))
		}
		clientResponseFields.AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		serverRequestLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, serverRequestLogLine.lvl)
		assert.Equal(s.T(), "request payload logged as grpc.request.content field", serverRequestLogLine.msg)
		serverRequestFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverRequestLogLine.fields, method, typ)
		serverRequestFields.AssertNextField(s.T(), "grpc.request.value", "something").
			AssertNextFieldNotEmpty(s.T(), "peer.address").
			AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
			AssertNextFieldNotEmpty(s.T(), "grpc.recv.duration").
			AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").
			AssertNextField(s.T(), "grpc.request.content", `{"value":"something","sleepTimeMs":9999}`).AssertNoMoreTags(s.T())
	}
	curr += repetitions
	for i := curr; i < curr+repetitions; i++ {
		serverResponseLogLine := lines[i]
		assert.Equal(s.T(), logging.INFO, serverResponseLogLine.lvl)
		assert.Equal(s.T(), "response payload logged as grpc.response.content field", serverResponseLogLine.msg)
		serverResponseFields := assertStandardFields(s.T(), logging.KindServerFieldValue, serverResponseLogLine.fields, method, typ)
		serverResponseFields = serverResponseFields.AssertNextField(s.T(), "grpc.request.value", "something").
			AssertNextFieldNotEmpty(s.T(), "peer.address").
			AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
			AssertNextFieldNotEmpty(s.T(), "grpc.send.duration").
			AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline")
		if i-curr == 0 {
			serverResponseFields = serverResponseFields.AssertNextField(s.T(), "grpc.response.content", `{"Value":"something"}`)
		} else {
			serverResponseFields = serverResponseFields.AssertNextField(s.T(), "grpc.response.content", fmt.Sprintf(`{"Value":"something","counter":%v}`, i-curr))
		}
		serverResponseFields.AssertNoMoreTags(s.T())
	}
}

func (s *loggingPayloadSuite) TestPingError_LogsOnlyRequestsOnError() {
	_, err := s.Client.PingError(s.SimpleCtx(), &testpb.PingRequest{Value: "something", ErrorCodeReturned: uint32(4)})
	require.Error(s.T(), err, "there must be an error on an unsuccessful call")

	lines := s.logger.Lines()
	require.Len(s.T(), lines, 2) // Only client & server requests.

	clientRequestLogLine := lines[0]
	assert.Equal(s.T(), logging.INFO, clientRequestLogLine.lvl)
	assert.Equal(s.T(), "request payload logged as grpc.request.content field", clientRequestLogLine.msg)
	clientRequestFields := assertStandardFields(s.T(), logging.KindClientFieldValue, clientRequestLogLine.fields, "PingError", interceptors.Unary)
	clientRequestFields.AssertNextFieldNotEmpty(s.T(), "grpc.start_time").
		AssertNextFieldNotEmpty(s.T(), "grpc.send.duration").
		AssertNextField(s.T(), "grpc.request.content", `{"value":"something","errorCodeReturned":4}`).
		AssertNextFieldNotEmpty(s.T(), "grpc.request.deadline").AssertNoMoreTags(s.T())
}

func (s *loggingPayloadSuite) TestPingStream_LogsAllRequestsAndResponses() {
	messagesExpected := 20
	stream, err := s.Client.PingStream(s.SimpleCtx())

	require.NoError(s.T(), err, "no error on stream creation")
	for i := 0; i < messagesExpected; i++ {
		require.NoError(s.T(), stream.Send(goodPing), "sending must succeed")

		pong := &testpb.PingResponse{}
		err := stream.RecvMsg(pong)
		require.NoError(s.T(), err, "no error on receive")
	}
	require.NoError(s.T(), stream.CloseSend(), "no error on send stream")

	lines := s.logger.Lines()
	require.Len(s.T(), lines, 4*messagesExpected)
	s.assertPayloadLogLinesForMessage(lines, "PingStream", interceptors.BidiStream)
}
