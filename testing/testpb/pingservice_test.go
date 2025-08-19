// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package testpb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestPingServiceOnWire(t *testing.T) {
	stopped := make(chan error)
	serverListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "must be able to allocate a port for serverListener")

	server := grpc.NewServer()
	RegisterTestServiceServer(server, &TestPingService{})

	go func() {
		defer close(stopped)
		stopped <- server.Serve(serverListener)
	}()
	defer func() {
		server.Stop()
		<-stopped
	}()

	// This is the point where we hook up the interceptor.
	clientConn, err := grpc.NewClient(
		serverListener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err, "must not error on client Dial")

	testClient := NewTestServiceClient(clientConn)
	select {
	case clientConnErr := <-stopped:
		t.Fatal("gRPC server stopped prematurely", clientConnErr)
	default:
	}

	r, err := testClient.PingEmpty(context.Background(), &PingEmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, r)

	r2, err := testClient.Ping(context.Background(), &PingRequest{Value: "24"})
	require.NoError(t, err)
	require.Equal(t, "24", r2.Value)
	require.Equal(t, int32(0), r2.Counter)

	_, err = testClient.PingError(context.Background(), &PingErrorRequest{
		ErrorCodeReturned: uint32(codes.Internal),
		Value:             "24",
	})
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))

	l, err := testClient.PingList(context.Background(), &PingListRequest{Value: "24"})
	require.NoError(t, err)
	for i := 0; i < ListResponseCount; i++ {
		r, receiveError := l.Recv()
		require.NoError(t, receiveError)
		require.Equal(t, "24", r.Value)
		require.Equal(t, int32(i), r.Counter)
	}

	s, err := testClient.PingStream(context.Background())
	require.NoError(t, err)
	for i := 0; i < ListResponseCount; i++ {
		require.NoError(t, s.Send(&PingStreamRequest{Value: fmt.Sprintf("%v", i)}))

		r, err := s.Recv()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%v", i), r.Value)
		require.Equal(t, int32(i), r.Counter)
	}

	select {
	case err := <-stopped:
		t.Fatal("gRPC server stopped prematurely", err)
	default:
	}
}

func TestTestServicePing_PingError(t *testing.T) {
	testCases := map[string]struct {
		request   *PingErrorRequest
		err       error
		unwrapped error
		msg       string
	}{
		"NotFound": {
			request:   &PingErrorRequest{ErrorCodeReturned: uint32(codes.NotFound), Value: "not found"},
			err:       &wrappedErrFields{wrappedErr: status.Error(codes.NotFound, "Userspace error"), fields: []any{"error-field", "plop"}},
			unwrapped: status.Error(codes.NotFound, "Userspace error"),
			msg:       "rpc error: code = NotFound desc = Userspace error",
		},
		"OK": {
			request:   &PingErrorRequest{ErrorCodeReturned: uint32(codes.OK), Value: "ok"},
			err:       &wrappedErrFields{wrappedErr: nil, fields: []any{"error-field", "plop"}},
			unwrapped: nil,
			msg:       "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			svc := &TestPingService{}

			_, err := svc.PingError(context.Background(), testCase.request)
			require.Equal(t, testCase.err, err)

			var we *wrappedErrFields
			ok := errors.As(err, &we)
			require.True(t, ok)

			assert.Equal(t, testCase.unwrapped, we.Unwrap())
			assert.Equal(t, testCase.msg, we.Error())
		})
	}
}
