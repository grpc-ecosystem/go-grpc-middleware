// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package timeout_test

import (
	"context"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type TimeoutTestServiceServer struct {
	sleepTime time.Duration
	testpb.TestPingService
}

func (t *TimeoutTestServiceServer) Ping(ctx context.Context, req *testpb.PingRequest) (*testpb.PingResponse, error) {
	if t.sleepTime > 0 {
		time.Sleep(t.sleepTime)
	}
	return t.TestPingService.Ping(ctx, req)
}

func TestTimeoutUnaryClientInterceptor(t *testing.T) {
	server := &TimeoutTestServiceServer{}

	its := &testpb.InterceptorTestSuite{
		ClientOpts: []grpc.DialOption{
			grpc.WithUnaryInterceptor(timeout.UnaryClientInterceptor(100 * time.Millisecond)),
		},
		TestService: server,
	}
	its.SetT(t)
	its.SetupSuite()
	defer its.TearDownSuite()

	// This call will take 0/100ms for respond, so the client timeout NOT exceed.
	resp, err := its.Client.Ping(context.TODO(), &testpb.PingRequest{Value: "default_response_value"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "default_response_value", resp.Value)

	// server will sleep 300ms before respond
	server.sleepTime = 300 * time.Millisecond

	// This call will take 300/100ms for respond, so the client timeout exceed.
	resp2, err2 := its.Client.Ping(context.TODO(), &testpb.PingRequest{})
	assert.Nil(t, resp2)
	assert.EqualError(t, err2, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
}
