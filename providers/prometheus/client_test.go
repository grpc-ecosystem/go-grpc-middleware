// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package prometheus

import (
	"io"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClientInterceptorSuite(t *testing.T) {
	c := NewClientMetrics(WithClientHandlingTimeHistogram())
	suite.Run(t, &ClientInterceptorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &testpb.TestPingService{},
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(c.UnaryClientInterceptor()),
				grpc.WithStreamInterceptor(c.StreamClientInterceptor()),
			},
		},
		clientMetrics: c,
	})
}

type ClientInterceptorTestSuite struct {
	*testpb.InterceptorTestSuite
	clientMetrics *ClientMetrics
}

func (s *ClientInterceptorTestSuite) SetupTest() {
	s.clientMetrics.clientStartedCounter.Reset()
	s.clientMetrics.clientHandledCounter.Reset()
	s.clientMetrics.clientHandledHistogram.Reset()
	s.clientMetrics.clientStreamMsgReceived.Reset()
	s.clientMetrics.clientStreamMsgSent.Reset()
}

func (s *ClientInterceptorTestSuite) TestUnaryIncrementsMetrics() {
	_, err := s.Client.PingEmpty(s.SimpleCtx(), &testpb.PingEmptyRequest{})
	require.NoError(s.T(), err)

	requireValue(s.T(), 1, s.clientMetrics.clientStartedCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingEmpty"))
	requireValue(s.T(), 1, s.clientMetrics.clientHandledCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingEmpty", "OK"))
	requireValueHistCount(s.T(), 1, s.clientMetrics.clientHandledHistogram.WithLabelValues("unary", testpb.TestServiceFullName, "PingEmpty"))

	_, err = s.Client.PingError(s.SimpleCtx(), &testpb.PingErrorRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)})
	require.Error(s.T(), err)
	requireValue(s.T(), 1, s.clientMetrics.clientStartedCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingError"))
	requireValue(s.T(), 1, s.clientMetrics.clientHandledCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingError", "FailedPrecondition"))
	requireValueHistCount(s.T(), 1, s.clientMetrics.clientHandledHistogram.WithLabelValues("unary", testpb.TestServiceFullName, "PingError"))
}

func (s *ClientInterceptorTestSuite) TestStartedStreamingIncrementsStarted() {
	_, err := s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{})
	require.NoError(s.T(), err)
	requireValue(s.T(), 1, s.clientMetrics.clientStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))

	_, err = s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)})
	require.NoError(s.T(), err, "PingList must not fail immediately")
	requireValue(s.T(), 2, s.clientMetrics.clientStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
}

func (s *ClientInterceptorTestSuite) TestStreamingIncrementsMetrics() {
	ss, err := s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{})
	require.NoError(s.T(), err)
	// Do a read, just for kicks.
	count := 0
	for {
		_, err = ss.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading pingList shouldn't fail")
		count++
	}
	require.EqualValues(s.T(), testpb.ListResponseCount, count, "Number of received msg on the wire must match")

	requireValue(s.T(), 1, s.clientMetrics.clientStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValue(s.T(), 1, s.clientMetrics.clientHandledCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList", "OK"))
	requireValue(s.T(), testpb.ListResponseCount+1 /* + EOF */, s.clientMetrics.clientStreamMsgReceived.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValue(s.T(), 1, s.clientMetrics.clientStreamMsgSent.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValueHistCount(s.T(), 1, s.clientMetrics.clientHandledHistogram.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))

	ss, err = s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)})
	require.NoError(s.T(), err, "PingList must not fail immediately")

	// Do a read, just to progate errors.
	_, err = ss.Recv()
	st, _ := status.FromError(err)
	require.Equal(s.T(), codes.FailedPrecondition, st.Code(), "Recv must return FailedPrecondition, otherwise the test is wrong")

	requireValue(s.T(), 2, s.clientMetrics.clientStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValue(s.T(), 1, s.clientMetrics.clientHandledCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList", "FailedPrecondition"))
	requireValueHistCount(s.T(), 2, s.clientMetrics.clientHandledHistogram.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
}

func (s *ClientInterceptorTestSuite) TestWithSubsystem() {
	counterOpts := []CounterOption{
		WithSubsystem("subsystem1"),
	}
	histOpts := []HistogramOption{
		WithHistogramSubsystem("subsystem1"),
	}
	clientCounterOpts := WithClientCounterOptions(counterOpts...)
	clientMetrics := NewClientMetrics(clientCounterOpts, WithClientHandlingTimeHistogram(histOpts...))

	requireSubsystemName(s.T(), "subsystem1", clientMetrics.clientStartedCounter.WithLabelValues("unary", testpb.TestServiceFullName, "dummy"))
	requireHistSubsystemName(s.T(), "subsystem1", clientMetrics.clientHandledHistogram.WithLabelValues("unary", testpb.TestServiceFullName, "dummy"))
}
