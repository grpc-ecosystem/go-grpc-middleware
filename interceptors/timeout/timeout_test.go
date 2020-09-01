package timeout_test

import (
	"context"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type TimeoutTestServiceServer struct {
	sleepTime time.Duration
	grpctesting.TestPingService
}

func (t *TimeoutTestServiceServer) PingEmpty(ctx context.Context, req *testpb.Empty) (*testpb.PingResponse, error) {
	time.Sleep(t.sleepTime)
	return &testpb.PingResponse{Value: "my_fake_value"}, nil
}

func TestTimeoutUnaryClientInterceptor(t *testing.T) {
	server := &TimeoutTestServiceServer{sleepTime: 1 * time.Millisecond}

	its := &grpctesting.InterceptorTestSuite{
		ClientOpts: []grpc.DialOption{
			grpc.WithUnaryInterceptor(timeout.TimeoutUnaryClientInterceptor(20 * time.Millisecond)),
		},
		TestService: server,
	}
	its.Suite.SetT(t)
	its.SetupSuite()
	defer its.TearDownSuite()

	resp, err := its.Client.PingEmpty(its.SimpleCtx(), &testpb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "my_fake_value", resp.Value)

	server.sleepTime = 30 * time.Millisecond
	resp2, err2 := its.Client.PingEmpty(its.SimpleCtx(), &testpb.Empty{})
	assert.Nil(t, resp2)
	assert.EqualError(t, err2, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")
}
