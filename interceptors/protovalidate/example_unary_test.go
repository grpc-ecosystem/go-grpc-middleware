// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate_test

import (
	"context"
	"net"

	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	testvalidatev1 "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UnaryService struct {
	testvalidatev1.TestValidateServiceServer
}

func (s *UnaryService) Send(_ context.Context, _ *testvalidatev1.SendRequest) (*testvalidatev1.SendResponse, error) {
	return &testvalidatev1.SendResponse{}, nil
}

func ExampleUnaryServerInterceptor() {
	validator, err := protovalidate.New()
	if err != nil {
		panic(err) // only for example purposes
	}

	var (
		srv = grpc.NewServer(
			grpc.UnaryInterceptor(
				protovalidate_middleware.UnaryServerInterceptor(validator,
					protovalidate_middleware.WithIgnoreMessages(
						(&testvalidatev1.SendRequest{}).ProtoReflect().Type(),
					),
					protovalidate_middleware.WithErrorConverter(
						func(err error) error {
							return status.Error(codes.InvalidArgument, err.Error())
						},
					),
				),
			),
		)
		svc = &UnaryService{}
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
