package filter_test

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/filter"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type someService struct {
	grpc_testing.TestPingService
	pingCalled       bool
	pingEmptyCalled  bool
	pingStreamCalled bool
}

func (s *someService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	s.pingCalled = true
	return s.TestPingService.Ping(ctx, ping)
}

func (s *someService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	s.pingEmptyCalled = true
	return s.TestPingService.PingEmpty(ctx, empty)
}

func (s *someService) PingStream(stream pb_testproto.TestService_PingStreamServer) error {
	s.pingStreamCalled = true
	return s.TestPingService.PingStream(stream)
}

type noopUnaryServerInterceptor struct {
	called bool
}

func (i *noopUnaryServerInterceptor) intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	i.called = true
	return handler(ctx, req)
}

type noopStreamServerInterceptor struct {
	called bool
}

func (i *noopStreamServerInterceptor) intercept(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	i.called = true
	return handler(srv, ss)
}

func TestServerMethods(t *testing.T) {
	service := &someService{
		TestPingService: grpc_testing.TestPingService{T: t},
	}
	si := &noopStreamServerInterceptor{}
	ui := &noopUnaryServerInterceptor{}
	suite.Run(t, &FilterSuite{
		srv: service,
		si:  si,
		ui:  ui,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: service,
			/*
				ClientOpts: []grpc.DialOption{
					grpc.WithStreamInterceptor(filter.StreamClientMethod()),
					grpc.WithUnaryInterceptor(filter.UnaryClientMethod()),
				},
			*/
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(filter.UnaryServerMethods(ui.intercept, true, "/mwitkow.testproto.TestService/Ping")),
				grpc.StreamInterceptor(filter.StreamServerMethods(si.intercept, true, "/mwitkow.testproto.TestService/PingStream")),
			},
		},
	})
}

type FilterSuite struct {
	*grpc_testing.InterceptorTestSuite
	srv *someService
	si  *noopStreamServerInterceptor
	ui  *noopUnaryServerInterceptor
}

func (s *FilterSuite) SetupTest() {
	s.srv.pingCalled = false
	s.srv.pingEmptyCalled = false
	s.srv.pingStreamCalled = false
	s.si.called = false
	s.ui.called = false
}

func (s *FilterSuite) TestUnary_CallAllowedUnaryMethod() {
	res, err := s.Client.Ping(s.SimpleCtx(), &pb_testproto.PingRequest{Value: "hello"})
	require.NoError(s.T(), err)
	require.Equal(s.T(), res.Value, "hello")
	require.True(s.T(), s.srv.pingCalled)
	require.False(s.T(), s.srv.pingEmptyCalled)
	require.False(s.T(), s.srv.pingStreamCalled)
	require.True(s.T(), s.ui.called) // allowed
	require.False(s.T(), s.si.called)
}

func (s *FilterSuite) TestUnary_CallDisallowedUnaryMethod() {
	_, err := s.Client.PingEmpty(s.SimpleCtx(), &pb_testproto.Empty{})
	require.NoError(s.T(), err)
	require.False(s.T(), s.srv.pingCalled)
	require.True(s.T(), s.srv.pingEmptyCalled)
	require.False(s.T(), s.srv.pingStreamCalled)
	require.False(s.T(), s.ui.called) // disallowed
	require.False(s.T(), s.si.called)
}
