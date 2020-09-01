package timeout_test

import (
	"context"
	"fmt"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"google.golang.org/grpc"
)

// Initialization shows an initialization sequence with a custom client request timeout.
func Example_initialization() {
	clientConn, err := grpc.Dial(
		"ServerAddr",
		grpc.WithUnaryInterceptor(
			// Set your client request timeout
			timeout.TimeoutUnaryClientInterceptor(20*time.Millisecond),
		),
	)

	// Handle connection error
	if err != nil {
		panic(err)
	}

	// Initialize your grpc service with connection
	testServiceClient := testpb.NewTestServiceClient(clientConn)
	resp, err := testServiceClient.PingEmpty(context.TODO(), &testpb.Empty{})

	// Handle request error
	if err != nil {
		panic(err)
	}

	// Use grpc response value
	fmt.Println(resp.Value)
}
