// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog_test

import (
	"io"
	"io/ioutil"
	"runtime"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	nullLogger = grpclog.NewLoggerV2WithVerbosity(ioutil.Discard, ioutil.Discard, ioutil.Discard, int(ctx_glog.DebugLevel))
)

func TestGLogPayloadSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }
	b := newGLogBaseSuite(t)
	b.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_glog.PayloadUnaryClientInterceptor(ctx_glog.NewEntry(b.logger), alwaysLoggingDeciderClient)),
		grpc.WithStreamInterceptor(grpc_glog.PayloadStreamClientInterceptor(ctx_glog.NewEntry(b.logger), alwaysLoggingDeciderClient)),
	}
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_glog.StreamServerInterceptor(ctx_glog.NewEntry(nullLogger)),
			grpc_glog.PayloadStreamServerInterceptor(ctx_glog.NewEntry(b.logger), alwaysLoggingDeciderServer)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_glog.UnaryServerInterceptor(ctx_glog.NewEntry(nullLogger)),
			grpc_glog.PayloadUnaryServerInterceptor(ctx_glog.NewEntry(b.logger), alwaysLoggingDeciderServer)),
	}
	suite.Run(t, &glogPayloadSuite{b})
}

type glogPayloadSuite struct {
	*glogBaseSuite
}

func (s *glogPayloadSuite) getServerAndClientMessages(expectedServer int, expectedClient int) (serverMsgs []map[string]interface{}, clientMsgs []map[string]interface{}) {
	msgs := s.getOutputJSONs()
	for _, m := range msgs {
		if m["span.kind"] == "server" {
			serverMsgs = append(serverMsgs, m)
		} else if m["span.kind"] == "client" {
			clientMsgs = append(clientMsgs, m)
		}
	}

	require.Len(s.T(), serverMsgs, expectedServer, "must match expected number of server log messages")
	require.Len(s.T(), clientMsgs, expectedClient, "must match expected number of client log messages")
	return serverMsgs, clientMsgs
}

func (s *glogPayloadSuite) TestPing_LogsBothRequestAndResponse() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "there must be not be an on a successful call")
	serverMsgs, clientMsgs := s.getServerAndClientMessages(2, 2)

	for _, m := range append(serverMsgs, clientMsgs...) {
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
		assert.Equal(s.T(), m["grpc.method"], "Ping", "all lines must contain the correct method name")
		assert.Equal(s.T(), m["level"], "info", "all lines must contain method name")
	}

	serverReq, serverResp := serverMsgs[0], serverMsgs[1]
	clientReq, clientResp := clientMsgs[0], clientMsgs[1]
	assert.Contains(s.T(), clientReq, "grpc.request.content", "request payload must be logged in a structured way")
	assert.Contains(s.T(), serverReq, "grpc.request.content", "request payload must be logged in a structured way")
	assert.Contains(s.T(), clientResp, "grpc.response.content", "response payload must be logged in a structured way")
	assert.Contains(s.T(), serverResp, "grpc.response.content", "response payload must be logged in a structured way")
}

func (s *glogPayloadSuite) TestPingError_LogsOnlyRequestsOnError() {
	_, err := s.Client.PingError(s.SimpleCtx(), &pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(4)})
	require.Error(s.T(), err, "there must be not be an error on a successful call")

	serverMsgs, clientMsgs := s.getServerAndClientMessages(1, 1)
	for _, m := range append(serverMsgs, clientMsgs...) {
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
		assert.Equal(s.T(), m["grpc.method"], "PingError", "all lines must contain the correct method name")
		assert.Equal(s.T(), m["level"], "info", "all lines must be logged at info level")
	}

	assert.Contains(s.T(), clientMsgs[0], "grpc.request.content", "request payload must be logged by the client")
	assert.Contains(s.T(), serverMsgs[0], "grpc.request.content", "request payload must be logged by the server")
}

func (s *glogPayloadSuite) TestPingStream_LogsAllRequestsAndResponses() {
	messagesExpected := 20
	stream, err := s.Client.PingStream(s.SimpleCtx())
	require.NoError(s.T(), err, "no error on stream creation")
	for i := 0; i < messagesExpected; i++ {
		require.NoError(s.T(), stream.Send(goodPing), "sending must succeed")
	}
	require.NoError(s.T(), stream.CloseSend(), "no error on close of stream")

	for {
		pong := &pb_testproto.PingResponse{}
		err := stream.RecvMsg(pong)
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "no error on receive")
	}
	serverMsgs, clientMsgs := s.getServerAndClientMessages(2*messagesExpected, 2*messagesExpected)
	for _, m := range append(serverMsgs, clientMsgs...) {
		assert.Equal(s.T(), m["grpc.service"], "mwitkow.testproto.TestService", "all lines must contain the correct service name")
		assert.Equal(s.T(), m["grpc.method"], "PingStream", "all lines must contain the correct method name")
		assert.Equal(s.T(), m["level"], "info", "all lines must be at info log level")

		content := m["grpc.request.content"] != nil || m["grpc.response.content"] != nil
		assert.True(s.T(), content, "all messages must contain a payload")
	}
}
