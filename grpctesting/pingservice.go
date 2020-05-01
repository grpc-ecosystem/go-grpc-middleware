// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
Package `grpc_testing` provides helper functions for testing validators in this package.
*/

package grpctesting

import (
	"context"
	"io"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// DefaultPongValue is the default value used.
	DefaultResponseValue = "default_response_value"
	// ListResponseCount is the expected number of responses to PingList
	ListResponseCount = 100
)

type TestPingService struct {
	T *testing.T
}

func (s *TestPingService) PingEmpty(_ context.Context, _ *testpb.Empty) (*testpb.PingResponse, error) {
	return &testpb.PingResponse{Value: DefaultResponseValue, Counter: 0}, nil
}

func (s *TestPingService) Ping(_ context.Context, ping *testpb.PingRequest) (*testpb.PingResponse, error) {
	// Send user trailers and headers.
	return &testpb.PingResponse{Value: ping.Value, Counter: 0}, nil
}

func (s *TestPingService) PingError(_ context.Context, ping *testpb.PingRequest) (*testpb.Empty, error) {
	code := codes.Code(ping.ErrorCodeReturned)
	return nil, status.Errorf(code, "Userspace error.")
}

func (s *TestPingService) PingList(ping *testpb.PingRequest, stream testpb.TestService_PingListServer) error {
	if ping.ErrorCodeReturned != 0 {
		return status.Errorf(codes.Code(ping.ErrorCodeReturned), "foobar")
	}
	// Send user trailers and headers.
	for i := 0; i < ListResponseCount; i++ {
		stream.Send(&testpb.PingResponse{Value: ping.Value, Counter: int32(i)})
	}
	return nil
}

func (s *TestPingService) PingStream(stream testpb.TestService_PingStreamServer) error {
	count := 0
	for true {
		ping, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		stream.Send(&testpb.PingResponse{Value: ping.Value, Counter: int32(count)})
		count += 1
	}
	return nil
}
