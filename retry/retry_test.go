// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_auth_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"sync"

	"github.com/mwitkow/go-grpc-middleware/testing"
	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

var (
	commonAuthToken   = "some_good_token"
	overrideAuthToken = "override_token"

	authedMarker = "some_context_marker"
	goodPing     = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type failingService struct {
	pb_testproto.TestServiceServer
	reqCounter uint
	reqModulo  uint
	reqError   codes.Code
	mu         sync.Mutex
}

func (s *failingService) resetFailingConfiguration(modulo uint, errorCode codes.Code) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reqCounter = 0
	s.reqModulo = modulo
	s.reqError = errorCode
}

func (s *failingService) maybeFailRequest() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reqCounter += 1
	if (s.reqModulo > 0) && (s.reqCounter%s.reqModulo == 0) {
		return nil
	}
	return grpc.Errorf(s.reqError, "maybeFailRequest: failing it")
}

func (s *failingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	if err := s.maybeFailRequest(); err != nil {
		return nil, err
	}
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *failingService) PingStream(stream pb_testproto.TestService_PingStreamServer) error {
	if err := s.maybeFailRequest(); err != nil {
		return err
	}
	return s.TestServiceServer.PingStream(stream)
}

func TestRetrySuite(t *testing.T) {
	service := &failingService{
		TestServiceServer: &grpc_testing.TestPingService{T: t},
	}
	s := &RetrySuite{
		srv: service,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: service,
			//ServerOpts: []grpc.ServerOption{
			//	grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
			//	grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)),
			//},
		},
	}
	suite.Run(t, s)
}

type RetrySuite struct {
	*grpc_testing.InterceptorTestSuite
	srv *failingService
}

func (s *RetrySuite) SetupTest() {
	s.srv.resetFailingConfiguration( /* don't fail */ 0, codes.OK)
}

func (s *RetrySuite) TestUnary_FailsOnNonRetriableError() {
	s.srv.resetFailingConfiguration(5, codes.Internal)
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.Error(s.T(), err, "error must occur from the failing service")
	require.Equal(s.T(), codes.Internal, grpc.Code(err), "failure code must come from retrier")
}

func (s *RetrySuite) TestStream_FailsOnNonRetriableError() {
	s.srv.resetFailingConfiguration(5, codes.Internal)
	stream, err := s.Client.PingStream(s.SimpleCtx())
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for i := 0; i < 5; i++ { // send some data, not too much
		err = stream.Send(goodPing)
		require.NoError(s.T(), err, "should not fail on sending")
	}
	_, err = stream.Recv()
	require.Error(s.T(), err, "error must occur from the failing service")
	require.Equal(s.T(), codes.Internal, grpc.Code(err), "failure code must come from retrier")
}
