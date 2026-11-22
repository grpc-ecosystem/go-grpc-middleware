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

type noopUnaryClientInterceptor struct {
	called bool
}

func (i *noopUnaryClientInterceptor) intercept(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	i.called = true
	return invoker(ctx, method, req, reply, cc, opts...)
}

type noopStreamClientInterceptor struct {
	called bool
}

func (i *noopStreamClientInterceptor) intercept(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	i.called = true
	return streamer(ctx, desc, cc, method, opts...)
}

func TestClientMethods(t *testing.T) {
	service := &someService{
		TestPingService: grpc_testing.TestPingService{T: t},
	}
	si := &noopStreamClientInterceptor{}
	ui := &noopUnaryClientInterceptor{}
	suite.Run(t, &ClientFilterSuite{
		srv: service,
		si:  si,
		ui:  ui,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: service,
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(filter.UnaryClientMethods(ui.intercept, true, "/mwitkow.testproto.TestService/Ping")),
				grpc.WithStreamInterceptor(filter.StreamClientMethods(si.intercept, true, "/mwitkow.testproto.TestService/PingStream")),
			},
		},
	})
}

type ClientFilterSuite struct {
	*grpc_testing.InterceptorTestSuite
	srv *someService
	si  *noopStreamClientInterceptor
	ui  *noopUnaryClientInterceptor
}

func (s *ClientFilterSuite) SetupTest() {
	s.srv.pingCalled = false
	s.srv.pingEmptyCalled = false
	s.srv.pingStreamCalled = false
	s.si.called = false
	s.ui.called = false
}

func (s *ClientFilterSuite) TestUnary_CallAllowedUnaryMethod() {
	res, err := s.Client.Ping(s.SimpleCtx(), &pb_testproto.PingRequest{Value: "hello"})
	require.NoError(s.T(), err)
	require.Equal(s.T(), res.Value, "hello")
	require.True(s.T(), s.srv.pingCalled)
	require.False(s.T(), s.srv.pingEmptyCalled)
	require.False(s.T(), s.srv.pingStreamCalled)
	require.True(s.T(), s.ui.called) // allowed
	require.False(s.T(), s.si.called)
}

func (s *ClientFilterSuite) TestUnary_CallDisallowedUnaryMethod() {
	_, err := s.Client.PingEmpty(s.SimpleCtx(), &pb_testproto.Empty{})
	require.NoError(s.T(), err)
	require.False(s.T(), s.srv.pingCalled)
	require.True(s.T(), s.srv.pingEmptyCalled)
	require.False(s.T(), s.srv.pingStreamCalled)
	require.False(s.T(), s.ui.called) // disallowed
	require.False(s.T(), s.si.called)
}
