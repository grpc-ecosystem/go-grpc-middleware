// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate

import (
	"context"
	"errors"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// UnaryServerInterceptor returns a new unary server interceptor that validates incoming messages.
func UnaryServerInterceptor(validator *protovalidate.Validator, opts ...Option) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		o := evaluateOpts(opts)
		switch msg := req.(type) {
		case proto.Message:
			if o.shouldIgnoreMessage(msg.ProtoReflect().Type()) {
				break
			}
			if err = validator.Validate(msg); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		default:
			return nil, errors.New("unsupported message type")
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that validates incoming messages.
func StreamServerInterceptor(validator *protovalidate.Validator, opts ...Option) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := stream.Context()

		wrapped := wrapServerStream(stream)
		wrapped.wrappedContext = ctx
		wrapped.validator = validator
		wrapped.options = evaluateOpts(opts)

		return handler(srv, wrapped)
	}
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	if err := w.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	msg := m.(proto.Message)
	if w.options.shouldIgnoreMessage(msg.ProtoReflect().Type()) {
		return nil
	}
	if err := w.validator.Validate(msg); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}

// wrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying context.
type wrappedServerStream struct {
	grpc.ServerStream
	// wrappedContext is the wrapper's own Context. You can assign it.
	wrappedContext context.Context

	validator *protovalidate.Validator
	options   *options
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *wrappedServerStream) Context() context.Context {
	return w.wrappedContext
}

// wrapServerStream returns a ServerStream that has the ability to overwrite context.
func wrapServerStream(stream grpc.ServerStream) *wrappedServerStream {
	return &wrappedServerStream{ServerStream: stream, wrappedContext: stream.Context()}
}
