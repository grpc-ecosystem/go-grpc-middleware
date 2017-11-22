package ctxlogger_zap_test

import (
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/zap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

func TestZapLoggingSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	b := newBaseZapSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			ctxlogger_zap.StreamServerInterceptor(b.log)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			ctxlogger_zap.UnaryServerInterceptor(b.log)),
	}
	suite.Run(t, &zapServerSuite{b})
}

type zapServerSuite struct {
	*zapBaseSuite
}

func (s *zapServerSuite) TestPing_WithCustomTags() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	msgs := s.getOutputJSONs()
	assert.Len(s.T(), msgs, 1, "single log statements should be logged")

	assert.Contains(s.T(), msgs[0], `"custom_tags.string": "something"`, "all lines must contain `custom_tags.string` set by AddFields")
	assert.Contains(s.T(), msgs[0], `"custom_tags.int": 1337`, "all lines must contain `custom_tags.int` set by AddFields")
	assert.Contains(s.T(), msgs[0], `"custom_field": "custom_value"`, "all lines must contain `custom_field` set by AddFields")
	assert.Contains(s.T(), msgs[0], `"msg": "some ping"`, "handler's message must contain user message")
}

func (s *zapServerSuite) TestPingList_WithCustomTags() {
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
	assert.Len(s.T(), msgs, 1, "single log statements should be logged")

	assert.Contains(s.T(), msgs[0], `"custom_tags.string": "something"`, "all lines must contain `custom_tags.string` set by AddFields")
	assert.Contains(s.T(), msgs[0], `"custom_tags.int": 1337`, "all lines must contain `custom_tags.int` set by AddFields")
	assert.Contains(s.T(), msgs[0], `"msg": "some pinglist"`, "handler's message must contain user message")
}
