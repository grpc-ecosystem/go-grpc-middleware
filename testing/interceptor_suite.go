// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_testing

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"net"
	"path"
	"runtime"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/testing/testcert"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	flagTls = flag.Bool("use_tls", true, "whether all gRPC middleware tests should use tls")
)

func getTestingCertsPath() string {
	_, callerPath, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(callerPath), "certs")
}

// InterceptorTestSuite is a testify/Suite that starts a gRPC PingService server and a client.
type InterceptorTestSuite struct {
	suite.Suite

	TestService pb_testproto.TestServiceServer
	ServerOpts  []grpc.ServerOption
	ClientOpts  []grpc.DialOption

	serverAddr     string
	ServerListener net.Listener
	Server         *grpc.Server
	clientConn     *grpc.ClientConn
	Client         pb_testproto.TestServiceClient

	restartServerWithDelayedStart chan time.Duration
	serverRunning                 chan bool
}

func (s *InterceptorTestSuite) SetupSuite() {
	s.restartServerWithDelayedStart = make(chan time.Duration)
	s.serverRunning = make(chan bool)

	s.serverAddr = "127.0.0.1:0"
	go func() {
		for {
			var err error
			s.ServerListener, err = net.Listen("tcp", s.serverAddr)
			if err != nil {
				s.T().Fatalf("unable to listen on address %s: %v", s.serverAddr, err)
			}
			s.serverAddr = s.ServerListener.Addr().String()
			require.NoError(s.T(), err, "must be able to allocate a port for serverListener")
			if *flagTls {
				cert, err := tls.X509KeyPair(testcert.KeyPairPEM())
				if err != nil {
					s.T().Fatalf("unable to load test TLS certificate: %v", err)
				}
				creds := credentials.NewServerTLSFromCert(&cert)
				s.ServerOpts = append(s.ServerOpts, grpc.Creds(creds))
			}
			// This is the point where we hook up the interceptor
			s.Server = grpc.NewServer(s.ServerOpts...)
			// Create a service of the instantiator hasn't provided one.
			if s.TestService == nil {
				s.TestService = &TestPingService{T: s.T()}
			}
			pb_testproto.RegisterTestServiceServer(s.Server, s.TestService)

			go func() {
				s.Server.Serve(s.ServerListener)
			}()
			if s.Client == nil {
				s.Client = s.NewClient(s.ClientOpts...)
			}

			s.serverRunning <- true

			d := <-s.restartServerWithDelayedStart
			s.Server.Stop()
			time.Sleep(d)
		}
	}()

	select {
	case <-s.serverRunning:
	case <-time.After(2 * time.Second):
		s.T().Fatal("server failed to start before deadline")
	}
}

func (s *InterceptorTestSuite) RestartServer(delayedStart time.Duration) <-chan bool {
	s.restartServerWithDelayedStart <- delayedStart
	time.Sleep(10 * time.Millisecond)
	return s.serverRunning
}

func (s *InterceptorTestSuite) NewClient(dialOpts ...grpc.DialOption) pb_testproto.TestServiceClient {
	newDialOpts := append(dialOpts, grpc.WithBlock())
	if *flagTls {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(testcert.CertPEM()) {
			s.T().Fatal("failed to append certificate")
		}
		creds := credentials.NewTLS(&tls.Config{ServerName: "localhost", RootCAs: cp})
		newDialOpts = append(newDialOpts, grpc.WithTransportCredentials(creds))
	} else {
		newDialOpts = append(newDialOpts, grpc.WithInsecure())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	clientConn, err := grpc.DialContext(ctx, s.ServerAddr(), newDialOpts...)
	require.NoError(s.T(), err, "must not error on client Dial")
	return pb_testproto.NewTestServiceClient(clientConn)
}

func (s *InterceptorTestSuite) ServerAddr() string {
	return s.serverAddr
}

func (s *InterceptorTestSuite) SimpleCtx() context.Context {
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	return ctx
}

func (s *InterceptorTestSuite) DeadlineCtx(deadline time.Time) context.Context {
	ctx, _ := context.WithDeadline(context.TODO(), deadline)
	return ctx
}

func (s *InterceptorTestSuite) TearDownSuite() {
	time.Sleep(10 * time.Millisecond)
	if s.ServerListener != nil {
		s.Server.GracefulStop()
		s.T().Logf("stopped grpc.Server at: %v", s.ServerAddr())
		s.ServerListener.Close()
	}
	if s.clientConn != nil {
		s.clientConn.Close()
	}
}
