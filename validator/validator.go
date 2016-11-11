// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
Package `grpc_validator` provides an easy way to hook protobuf message validation as a gRPC
interceptor across all your APIs.

It primarily meant to be used with https://github.com/mwitkow/go-proto-validators, which code-gen
assertions about allowed values from `.proto` files.

Basically this will invoke a .Validate() method on incoming message of the stream, if such method is
defined. If that method returns an error, an `INVALID_ARGUMENT` gRPC status code is returned.
*/

package grpc_validator

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type validator interface {
	Validate() error
}

// UnaryServerInterceptor returns a new unary server interceptors that validates incoming messages.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(validator); ok {
			if err := v.Validate(); err != nil {
				return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
			}
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptors that validates incoming messages.
// The validation happens on message receives.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{stream}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
}

func (s *recvWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if v, ok := m.(validator); ok {
		if err := v.Validate(); err != nil {
			return grpc.Errorf(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}
