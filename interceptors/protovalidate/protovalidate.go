// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate

import (
	"context"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func validateMessage(validator *protovalidate.Validator, o *options, req any) error {
	msg := req.(proto.Message)

	if o.shouldIgnoreMessage(msg.ProtoReflect().Type()) {
		return nil
	}

	if err := validator.Validate(msg); err != nil {
		return o.errorConverter(err)
	}

	return nil
}

// UnaryServerInterceptor returns a new unary server interceptor that validates incoming messages.
func UnaryServerInterceptor(validator *protovalidate.Validator, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateOpts(opts)

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if err := validateMessage(validator, o, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that validates incoming messages.
func StreamServerInterceptor(validator *protovalidate.Validator, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateOpts(opts)

	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		wrapped := wrapServerStream(stream, validator, o)

		return handler(srv, wrapped)
	}
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	if err := validateMessage(w.validator, w.options, m); err != nil {
		return err
	}

	return w.ServerStream.RecvMsg(m)
}

// wrappedServerStream is a thin wrapper around grpc.ServerStream that allows to validate messages.
type wrappedServerStream struct {
	grpc.ServerStream
	validator *protovalidate.Validator
	options   *options
}

// wrapServerStream returns a ServerStream that has the ability to validate messages.
func wrapServerStream(
	stream grpc.ServerStream,
	validator *protovalidate.Validator,
	options *options,
) *wrappedServerStream {
	return &wrappedServerStream{ServerStream: stream, validator: validator, options: options}
}
