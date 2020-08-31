// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_timeout_test

import (
	"context"
	"fmt"
	"time"

	mwitkow_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	grpc_timeout "github.com/grpc-ecosystem/go-grpc-middleware/timeout"
	"google.golang.org/grpc"
)

// Initialization shows an initialization sequence with a custom client request timeout.
func Example_initialization() {
	clientConn, err := grpc.Dial(
		"ServerAddr",
		grpc.WithUnaryInterceptor(
			// Set your client request timeout
			grpc_timeout.TimeoutUnaryClientInterceptor(20*time.Millisecond),
		),
	)

	// Handle connection error
	if err != nil {
		panic(err)
	}

	// Initialize your grpc service with connection
	testServiceClient := mwitkow_testproto.NewTestServiceClient(clientConn)
	resp, err := testServiceClient.PingEmpty(context.TODO(), &mwitkow_testproto.Empty{})

	// Handle request error
	if err != nil {
		panic(err)
	}

	// Use grpc response value
	fmt.Println(resp.Value)
}
