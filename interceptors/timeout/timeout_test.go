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
	if t.sleepTime > 0 {
		time.Sleep(t.sleepTime)
	}
	return t.TestPingService.PingEmpty(ctx, req)
}

func TestTimeoutUnaryClientInterceptor(t *testing.T) {
	server := &TimeoutTestServiceServer{}

	its := &grpctesting.InterceptorTestSuite{
		ClientOpts: []grpc.DialOption{
			grpc.WithUnaryInterceptor(timeout.TimeoutUnaryClientInterceptor(100 * time.Millisecond)),
		},
		TestService: server,
	}
	its.Suite.SetT(t)
	its.SetupSuite()
	defer its.TearDownSuite()

	// This call will take 0/100ms for respond, so the client timeout NOT exceed.
	resp, err := its.Client.PingEmpty(context.TODO(), &testpb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "default_response_value", resp.Value)

	// server will sleep 300ms before respond
	server.sleepTime = 300 * time.Millisecond

	// This call will take 300/100ms for respond, so the client timeout exceed.
	resp2, err2 := its.Client.PingEmpty(context.TODO(), &testpb.Empty{})
	assert.Nil(t, resp2)
	assert.EqualError(t, err2, "rpc error: code = DeadlineExceeded desc = context deadline exceeded")

	// This call will take 300/400ms for respond, so the client timeout NOT exceed.
	longerValidityContext, cancel := context.WithTimeout(context.TODO(), 400*time.Millisecond)
	defer cancel()
	resp3, err3 := its.Client.PingEmpty(longerValidityContext, &testpb.Empty{})
	assert.NoError(t, err3)
	assert.NotNil(t, resp3)
	assert.Equal(t, "default_response_value", resp.Value)
}