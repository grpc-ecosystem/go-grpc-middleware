// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package timeout_test

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

// Initialization shows an initialization sequence with a custom client request timeout.
func Example_initialization() {
	clientConn, err := grpc.Dial(
		"ServerAddr",
		grpc.WithUnaryInterceptor(
			// Set your client request timeout.
			timeout.UnaryClientInterceptor(20*time.Millisecond),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize your grpc service with connection.
	testServiceClient := testpb.NewTestServiceClient(clientConn)
	resp, err := testServiceClient.Ping(context.TODO(), &testpb.PingRequest{Value: "my_example_value"})
	if err != nil {
		log.Fatal(err)
	}

	// Use grpc response value.
	log.Println(resp.Value)
}
