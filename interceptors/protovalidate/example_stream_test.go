// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate_test

import (
	"net"

	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	testvalidatev1 "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate/v1"
	"google.golang.org/grpc"
)

type StreamService struct {
	testvalidatev1.TestValidateServiceServer
}

func (s *StreamService) SendStream(_ *testvalidatev1.SendStreamRequest, _ testvalidatev1.TestValidateService_SendStreamServer) error {
	return nil
}

func ExampleStreamServerInterceptor() {
	validator, err := protovalidate.New()
	if err != nil {
		panic(err) // only for example purposes
	}

	var (
		srv = grpc.NewServer(
			grpc.StreamInterceptor(
				protovalidate_middleware.StreamServerInterceptor(validator,
					protovalidate_middleware.WithIgnoreMessages(
						(&testvalidatev1.SendStreamRequest{}).ProtoReflect().Type(),
					)),
			),
		)
		svc = &StreamService{}
	)

	testvalidatev1.RegisterTestValidateServiceServer(srv, svc)

	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		panic(err) // only for example purposes
	}

	if err = srv.Serve(listener); err != nil {
		panic(err) // only for example purposes
	}
}
