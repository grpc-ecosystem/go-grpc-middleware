// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_testing

import (
	"net"
	"time"

	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// InterceptorTestSuite is a testify/Suite that starts a gRPC PingService server and a client.
type InterceptorTestSuite struct {
	suite.Suite

	TestService pb_testproto.TestServiceServer
	ServerOpts []grpc.ServerOption
	ClientOpts []grpc.DialOption

	ServerListener net.Listener
	Server         *grpc.Server
	clientConn     *grpc.ClientConn
	Client         pb_testproto.TestServiceClient
}

func (s *InterceptorTestSuite) SetupSuite() {
	var err error
	s.ServerListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	// This is the point where we hook up the interceptor
	s.Server = grpc.NewServer(s.ServerOpts...)
	// Crete a service of the instantiator hasn't provided one.
	if s.TestService == nil {
		s.TestService = &TestPingService{T: s.T()}
	}
	pb_testproto.RegisterTestServiceServer(s.Server, s.TestService)

	go func() {
		s.Server.Serve(s.ServerListener)
	}()
	s.Client = s.NewClient(s.ClientOpts...)
}

func (s *InterceptorTestSuite) NewClient(dialOpts ...grpc.DialOption) pb_testproto.TestServiceClient {
	newDialOpts := append(dialOpts, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	clientConn, err := grpc.Dial(s.ServerAddr(), newDialOpts...)
	require.NoError(s.T(), err, "must not error on client Dial")
	return pb_testproto.NewTestServiceClient(clientConn)
}

func (s *InterceptorTestSuite) ServerAddr() string {
	return s.ServerListener.Addr().String()
}

func (s *InterceptorTestSuite) SimpleCtx() context.Context {
	ctx, _ := context.WithTimeout(context.TODO(), 2 * time.Second)
	return ctx
}

func (s *InterceptorTestSuite) TearDownSuite() {
	if s.ServerListener != nil {
		s.Server.Stop()
		s.T().Logf("stopped grpc.Server at: %v", s.ServerAddr())
		s.ServerListener.Close()

	}
	if s.clientConn != nil {
		s.clientConn.Close()
	}
}
