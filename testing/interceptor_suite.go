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

	ServerOpts []grpc.ServerOption
	ClientOpts []grpc.DialOption

	ServerListener net.Listener
	Server         *grpc.Server
	clientConn     *grpc.ClientConn
	Client         pb_testproto.TestServiceClient
	ctx            context.Context
}

func (s *InterceptorTestSuite) SetupSuite() {
	var err error
	s.ServerListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	// This is the point where we hook up the interceptor
	s.Server = grpc.NewServer(s.ServerOpts...)
	pb_testproto.RegisterTestServiceServer(s.Server, &TestPingService{T: s.T()})

	go func() {
		s.Server.Serve(s.ServerListener)
	}()
	clientOpts := append(s.ClientOpts, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	s.clientConn, err = grpc.Dial(s.ServerAddr(), clientOpts...)
	require.NoError(s.T(), err, "must not error on client Dial")
	s.Client = pb_testproto.NewTestServiceClient(s.clientConn)

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
