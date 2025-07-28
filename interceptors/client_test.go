// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package interceptors

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type mockReport struct {
	CallMeta

	postCalls       []error
	postMsgSends    []error
	postMsgReceives []error
}

type mockReportable struct {
	m    sync.Mutex
	curr *mockReport

	reports []*mockReport
}

// Equal replaces require.Equal as google.golang.org/grpc/status errors are not easily comparable.
func (m *mockReportable) Equal(t *testing.T, expected []*mockReport) {
	t.Helper()

	require.Len(t, expected, len(m.reports))
	for i, e := range m.reports {
		require.Equal(t, expected[i].Typ, e.Typ, "%v", i)
		require.Equal(t, expected[i].Service, e.Service, "%v", i)
		require.Equal(t, expected[i].Method, e.Method, "%v", i)

		require.Len(t, expected[i].postCalls, len(e.postCalls), "%v", i)
		for k, err := range e.postCalls {
			if expected[i].postCalls[k] == nil {
				require.NoError(t, err)
				continue
			}
			require.EqualError(t, err, expected[i].postCalls[k].Error(), "%v %v", i, k)
		}
		require.Len(t, expected[i].postMsgSends, len(e.postMsgSends), "%v", i)
		for k, err := range e.postMsgSends {
			if expected[i].postMsgSends[k] == nil {
				require.NoError(t, err)
				continue
			}
			require.Equal(t, expected[i].postMsgSends[k].Error(), err.Error(), "%v %v", i, k)
		}
		require.Len(t, expected[i].postMsgReceives, len(e.postMsgReceives), "%v", i)
		for k, err := range e.postMsgReceives {
			if expected[i].postMsgReceives[k] == nil {
				require.NoError(t, err)
				continue
			}
			require.Equal(t, expected[i].postMsgReceives[k].Error(), err.Error(), "%v %v", i, k)
		}

	}
}

func (m *mockReportable) requireOneReportWithRetry(ctx context.Context, t *testing.T, expected *mockReport) {
	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for mockReport")
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
	// Even without reading, we should get initial mockReport.
	m.Equal(t, []*mockReport{expected})
}

func (m *mockReportable) PostCall(err error, _ time.Duration) {
	m.m.Lock()
	defer m.m.Unlock()
	m.curr.postCalls = append(m.curr.postCalls, err)
}

func (m *mockReportable) PostMsgSend(_ any, err error, _ time.Duration) {
	m.m.Lock()
	defer m.m.Unlock()
	m.curr.postMsgSends = append(m.curr.postMsgSends, err)
}

func (m *mockReportable) PostMsgReceive(_ any, err error, _ time.Duration) {
	m.m.Lock()
	defer m.m.Unlock()
	m.curr.postMsgReceives = append(m.curr.postMsgReceives, err)
}

func (m *mockReportable) ClientReporter(ctx context.Context, c CallMeta) (Reporter, context.Context) {
	m.curr = &mockReport{CallMeta: c}
	m.reports = append(m.reports, m.curr)
	return m, ctx
}

func (m *mockReportable) ServerReporter(ctx context.Context, c CallMeta) (Reporter, context.Context) {
	m.curr = &mockReport{CallMeta: c}
	m.reports = append(m.reports, m.curr)
	return m, ctx
}

func TestClientInterceptorSuite(t *testing.T) {
	suite.Run(t, &ClientInterceptorTestSuite{})
}

type ClientInterceptorTestSuite struct {
	suite.Suite

	serverListener net.Listener
	server         *grpc.Server
	clientConn     *grpc.ClientConn
	testClient     testpb.TestServiceClient
	ctx            context.Context
	cancel         context.CancelFunc

	mock *mockReportable

	stopped chan error
}

func (s *ClientInterceptorTestSuite) SetupSuite() {
	var err error
	s.stopped = make(chan error)
	s.mock = &mockReportable{}

	s.serverListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	s.server = grpc.NewServer()
	testpb.RegisterTestServiceServer(s.server, &testpb.TestPingService{})

	go func() {
		defer close(s.stopped)
		s.stopped <- s.server.Serve(s.serverListener)
	}()

	// This is the point where we hook up the interceptor.
	s.clientConn, err = grpc.NewClient(
		s.serverListener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor(s.mock)),
		grpc.WithStreamInterceptor(StreamClientInterceptor(s.mock)),
	)
	require.NoError(s.T(), err, "must not error on client Dial")
	s.testClient = testpb.NewTestServiceClient(s.clientConn)
}

func (s *ClientInterceptorTestSuite) SetupTest() {
	select {
	case err := <-s.stopped:
		s.T().Fatal("gRPC server stopped prematurely", err)
	default:
	}

	// Make all RPC calls last at most 2 sec, meaning all async issues or deadlock will not kill tests.
	s.ctx, s.cancel = context.WithTimeout(context.TODO(), 2*time.Second)
	s.mock.reports = s.mock.reports[:0]
}

func (s *ClientInterceptorTestSuite) TearDownSuite() {
	if s.serverListener != nil {
		s.server.Stop()
		s.T().Logf("stopped grpc.Server at: %v", s.serverListener.Addr().String())
		_ = s.serverListener.Close()
	}
	if s.clientConn != nil {
		_ = s.clientConn.Close()
	}
	<-s.stopped
}

func (s *ClientInterceptorTestSuite) TearDownTest() {
	s.cancel()
}

func (s *ClientInterceptorTestSuite) TestUnaryReporting() {
	_, err := s.testClient.PingEmpty(s.ctx, &testpb.PingEmptyRequest{}) // should return with code=OK
	require.NoError(s.T(), err)
	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:        CallMeta{Typ: Unary, Service: testpb.TestServiceFullName, Method: "PingEmpty"},
		postCalls:       []error{nil},
		postMsgReceives: []error{nil},
		postMsgSends:    []error{nil},
	}})
	s.mock.reports = s.mock.reports[:0] // Reset.

	_, err = s.testClient.PingError(s.ctx, &testpb.PingErrorRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	require.Error(s.T(), err)
	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:        CallMeta{Typ: Unary, Service: testpb.TestServiceFullName, Method: "PingError"},
		postCalls:       []error{status.Error(codes.FailedPrecondition, "Userspace error")},
		postMsgReceives: []error{status.Error(codes.FailedPrecondition, "Userspace error")},
		postMsgSends:    []error{nil},
	}})
}

func (s *ClientInterceptorTestSuite) TestStartedListReporting() {
	_, err := s.testClient.PingList(s.ctx, &testpb.PingListRequest{})
	require.NoError(s.T(), err)

	// Even without reading, we should get initial mockReport.
	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:     CallMeta{Typ: ServerStream, Service: testpb.TestServiceFullName, Method: "PingList"},
		postMsgSends: []error{nil},
	}})

	_, err = s.testClient.PingList(s.ctx, &testpb.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)})
	require.NoError(s.T(), err, "PingList must not fail immediately")

	// Even without reading, we should get initial mockReport.
	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:     CallMeta{Typ: ServerStream, Service: testpb.TestServiceFullName, Method: "PingList"},
		postMsgSends: []error{nil},
	}, {
		CallMeta:     CallMeta{Typ: ServerStream, Service: testpb.TestServiceFullName, Method: "PingList"},
		postMsgSends: []error{nil},
	}})
}

func (s *ClientInterceptorTestSuite) TestListReporting() {
	ss, err := s.testClient.PingList(s.ctx, &testpb.PingListRequest{})
	require.NoError(s.T(), err)

	// Do a read, just for kicks.
	count := 0
	for {
		_, err := ss.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(s.T(), err, "reading pingList shouldn't fail")
		count++
	}
	require.EqualValues(s.T(), testpb.ListResponseCount, count, "Number of received msg on the wire must match")

	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:        CallMeta{Typ: ServerStream, Service: testpb.TestServiceFullName, Method: "PingList"},
		postCalls:       []error{nil},
		postMsgReceives: append(make([]error, testpb.ListResponseCount), io.EOF),
		postMsgSends:    []error{nil},
	}})
	s.mock.reports = s.mock.reports[:0] // Reset.

	ss, err = s.testClient.PingList(s.ctx, &testpb.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)})
	require.NoError(s.T(), err, "PingList must not fail immediately")

	// Do a read, just to propagate errors.
	_, err = ss.Recv()
	require.Error(s.T(), err)
	st, _ := status.FromError(err)
	require.Equal(s.T(), codes.FailedPrecondition, st.Code(), "Recv must return FailedPrecondition, otherwise the test is wrong")

	// Next same.
	_, err = ss.Recv()
	require.Error(s.T(), err)
	st, _ = status.FromError(err)
	require.Equal(s.T(), codes.FailedPrecondition, st.Code(), "Recv must return FailedPrecondition, otherwise the test is wrong")

	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:        CallMeta{Typ: ServerStream, Service: testpb.TestServiceFullName, Method: "PingList"},
		postCalls:       []error{status.Error(codes.FailedPrecondition, "foobar"), status.Error(codes.FailedPrecondition, "foobar")},
		postMsgReceives: []error{status.Error(codes.FailedPrecondition, "foobar"), status.Error(codes.FailedPrecondition, "foobar")},
		postMsgSends:    []error{nil},
	}})
}

func (s *ClientInterceptorTestSuite) TestBiStreamingReporting() {
	ss, err := s.testClient.PingStream(s.ctx)
	require.NoError(s.T(), err)

	wg := sync.WaitGroup{}
	defer func() {
		_ = ss.CloseSend()
		wg.Wait()
	}()

	count := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for s.ctx.Err() == nil {

			_, err := ss.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if !s.Assert().NoError(err, "reading pingStream shouldn't fail") {
				break
			}
			count++
		}
	}()
	for i := 0; i < 100; i++ {
		require.NoError(s.T(), ss.Send(&testpb.PingStreamRequest{}), "sending shouldn't fail")
	}

	require.NoError(s.T(), ss.CloseSend())
	wg.Wait()

	require.EqualValues(s.T(), 100, count, "Number of received msg on the wire must match")
	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:        CallMeta{Typ: BidiStream, Service: testpb.TestServiceFullName, Method: "PingStream"},
		postCalls:       []error{nil},
		postMsgReceives: append(make([]error, 100), io.EOF),
		postMsgSends:    make([]error, 100),
	}})
}

func (s *ClientInterceptorTestSuite) TestClientStream() {
	ss, err := s.testClient.PingClientStream(s.ctx)
	require.NoError(s.T(), err)

	defer func() {
		_, _ = ss.CloseAndRecv()
	}()

	for i := 0; i < 100; i++ {
		require.NoError(s.T(), ss.Send(&testpb.PingClientStreamRequest{}), "sending shouldn't fail")
	}

	_, err = ss.CloseAndRecv()
	require.NoError(s.T(), err)

	s.mock.Equal(s.T(), []*mockReport{{
		CallMeta:        CallMeta{Typ: ClientStream, Service: testpb.TestServiceFullName, Method: "PingClientStream"},
		postCalls:       []error{nil},
		postMsgReceives: []error{nil},
		postMsgSends:    make([]error, 100),
	}})
}
