// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package interceptors

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
)

type mockServerReportable struct {
	m       sync.Mutex
	reports []*mockedReporter
}

func (m *mockServerReportable) ServerReporter(ctx context.Context, _ interface{}, typ GRPCType, serviceName string, methodName string) (Reporter, context.Context) {
	mock := &mockedReporter{m: &m.m, typ: typ, svcName: serviceName, methodName: methodName}
	m.m.Lock()
	defer m.m.Unlock()

	m.reports = append(m.reports, mock)
	return mock, ctx
}

func TestServerInterceptorSuite(t *testing.T) {
	suite.Run(t, &ServerInterceptorTestSuite{})
}

type ServerInterceptorTestSuite struct {
	suite.Suite

	serverListener net.Listener
	server         *grpc.Server
	clientConn     *grpc.ClientConn
	testClient     testpb.TestServiceClient
	ctx            context.Context
	cancel         context.CancelFunc

	mock *mockServerReportable
}

func (s *ServerInterceptorTestSuite) SetupSuite() {
	var err error

	s.mock = &mockServerReportable{}

	s.serverListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	// This is the point where we hook up the interceptor
	s.server = grpc.NewServer(
		grpc.StreamInterceptor(StreamServerInterceptor(s.mock)),
		grpc.UnaryInterceptor(UnaryServerInterceptor(s.mock)),
	)
	testpb.RegisterTestServiceServer(s.server, &grpctesting.TestPingService{T: s.T()})

	go func() {
		s.server.Serve(s.serverListener)
	}()

	s.clientConn, err = grpc.Dial(s.serverListener.Addr().String(), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	require.NoError(s.T(), err, "must not error on client Dial")
	s.testClient = testpb.NewTestServiceClient(s.clientConn)
}

func (s *ServerInterceptorTestSuite) SetupTest() {
	// Make all RPC calls last at most 2 sec, meaning all async issues or deadlock will not kill tests.
	s.ctx, s.cancel = context.WithTimeout(context.TODO(), 2*time.Second)

	s.mock.reports = s.mock.reports[:0]
}

func (s *ServerInterceptorTestSuite) TearDownSuite() {
	if s.serverListener != nil {
		s.server.Stop()
		s.T().Logf("stopped grpc.Server at: %v", s.serverListener.Addr().String())
		s.serverListener.Close()

	}
	if s.clientConn != nil {
		s.clientConn.Close()
	}
}

func (s *ServerInterceptorTestSuite) TearDownTest() {
	s.cancel()
}

func (s *ServerInterceptorTestSuite) TestUnaryReporting() {
	_, err := s.testClient.PingEmpty(s.ctx, &testpb.Empty{}) // should return with code=OK
	require.NoError(s.T(), err)
	require.Equal(s.T(), []*mockedReporter{{
		m:               &sync.Mutex{},
		typ:             Unary,
		svcName:         testpb.TestServiceFullName,
		methodName:      "PingEmpty",
		postCalls:       []error{nil},
		postMsgReceives: []error{nil},
		postMsgSends:    []error{nil},
	}}, s.mock.reports)
	s.mock.reports = s.mock.reports[:0] // Reset.

	_, err = s.testClient.PingError(s.ctx, &testpb.PingRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.Error(s.T(), err)
	require.Equal(s.T(), []*mockedReporter{{
		m:               &sync.Mutex{},
		typ:             Unary,
		svcName:         testpb.TestServiceFullName,
		methodName:      "PingError",
		postCalls:       []error{status.Errorf(codes.FailedPrecondition, "Userspace error.")},
		postMsgReceives: []error{nil},
		postMsgSends:    []error{status.Errorf(codes.FailedPrecondition, "Userspace error.")},
	}}, s.mock.reports)
}

func (s *ServerInterceptorTestSuite) TestStreamingReports() {
	ss, _ := s.testClient.PingList(s.ctx, &testpb.PingRequest{}) // should return with code=OK
	// Do a read, just for kicks.
	count := 0
	for {
		_, err := ss.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading pingList shouldn't fail")
		count++
	}
	require.EqualValues(s.T(), grpctesting.ListResponseCount, count, "Number of received msg on the wire must match")
	require.Equal(s.T(), []*mockedReporter{{
		m:               &sync.Mutex{},
		typ:             ServerStream,
		svcName:         testpb.TestServiceFullName,
		methodName:      "PingList",
		postCalls:       []error{nil},
		postMsgReceives: []error{nil},
		postMsgSends:    append(make([]error, grpctesting.ListResponseCount)),
	}}, s.mock.reports)
	s.mock.reports = s.mock.reports[:0] // Reset.

	_, err := s.testClient.PingList(s.ctx, &testpb.PingRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.NoError(s.T(), err, "PingList must not fail immediately")

	s.mock.requireOneReportWithRetry(s.ctx, s.T(), &mockedReporter{
		typ:             ServerStream,
		svcName:         testpb.TestServiceFullName,
		methodName:      "PingList",
		postCalls:       []error{status.Errorf(codes.FailedPrecondition, "foobar")},
		postMsgReceives: []error{nil},
	})
}

func (m *mockServerReportable) requireOneReportWithRetry(ctx context.Context, t *testing.T, expected *mockedReporter) {
	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for report")
		case <-time.After(200 * time.Millisecond):
		}

		m.m.Lock()
		if len(m.reports) == 0 {
			m.m.Unlock()
			continue
		}
		defer m.m.Unlock()
		break
	}
	expected.m = &m.m
	// Even without reading, we should get initial report.
	require.Equal(t, []*mockedReporter{expected}, m.reports)
}

func (s *ServerInterceptorTestSuite) TestBiStreamingReporting() {
	ss, err := s.testClient.PingStream(s.ctx)
	require.NoError(s.T(), err)
	wg := sync.WaitGroup{}

	defer func() {
		ss.CloseSend()
		wg.Wait()
	}()

	count := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if s.ctx.Err() != nil {
				break
			}
			_, err := ss.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(s.T(), err, "reading pingStream shouldn't fail")
			count++
		}
	}()
	for i := 0; i < 100; i++ {
		require.NoError(s.T(), ss.Send(&testpb.PingRequest{}), "sending shouldn't fail")
	}

	require.NoError(s.T(), ss.CloseSend())
	wg.Wait()

	require.EqualValues(s.T(), count, 100, "Number of received msg on the wire must match")

	require.Equal(s.T(), []*mockedReporter{{
		m:               &sync.Mutex{},
		typ:             BidiStream,
		svcName:         testpb.TestServiceFullName,
		methodName:      "PingStream",
		postCalls:       []error{nil},
		postMsgReceives: append(make([]error, 100), io.EOF),
		postMsgSends:    append(make([]error, 100)),
	}}, s.mock.reports)
}
