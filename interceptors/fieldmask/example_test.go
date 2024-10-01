// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package fieldmask

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"google.golang.org/grpc"
)

// Simple example of server initialization code.
func Example_serverConfig() {
	_ = grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor(DefaultFilterFunc)),
	)
}

// Simple example of server initialization code with fieldmask interceptor.
func Example_serverConfigWithAuthOverride() {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor(DefaultFilterFunc)),
	)
	testpb.RegisterTestServiceServer(server, &testpb.TestPingService{})
}
