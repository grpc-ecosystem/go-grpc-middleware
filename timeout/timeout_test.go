package grpc_timeout_test

import (
	"context"
	"testing"
	"time"

	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	mwitkow_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/grpc-ecosystem/go-grpc-middleware/timeout"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type TimeoutTestServiceServer struct {
	sleepTime time.Duration
	mwitkow_testproto.UnimplementedTestServiceServer
}

func (t *TimeoutTestServiceServer) PingEmpty(ctx context.Context, req *mwitkow_testproto.Empty) (*mwitkow_testproto.PingResponse, error) {
	time.Sleep(t.sleepTime)
	return &mwitkow_testproto.PingResponse{Value: "my_fake_value"}, nil
}

func TestTimeoutUnaryClientInterceptor(t *testing.T) {
	server := &TimeoutTestServiceServer{sleepTime: 1 * time.Millisecond}

	its := &grpc_testing.InterceptorTestSuite{
		ClientOpts: []grpc.DialOption{
			grpc.WithUnaryInterceptor(grpc_timeout.TimeoutUnaryClientInterceptor(20 * time.Millisecond)),
		},
		TestService: server,
	}
	its.Suite.SetT(t)
	its.SetupSuite()
	defer its.TearDownSuite()

	resp, err := its.Client.PingEmpty(its.SimpleCtx(), &mwitkow_testproto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "my_fake_value", resp.Value)

	server.sleepTime = 30 * time.Millisecond
	resp2, err2 := its.Client.PingEmpty(its.SimpleCtx(), &mwitkow_testproto.Empty{})
	assert.Nil(t, resp2)
	assert.EqualError(t, err2, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
}
