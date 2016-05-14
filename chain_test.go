// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_middleware

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	parentUnaryInfo  = &grpc.UnaryServerInfo{FullMethod: "SomeService.UnaryMethod"}
	parentStreamInfo = &grpc.StreamServerInfo{
		FullMethod:     "SomeService.StreamMethod",
		IsServerStream: true,
	}
	someValue     = 1
	parentContext = context.WithValue(context.TODO(), "parent", someValue)
)

func TestChainUnaryServer(t *testing.T) {
	input := "input"
	output := "output"

	first := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requireContextValue(t, ctx, "parent", "first interceptor must know the parent context value")
		require.Equal(t, parentUnaryInfo, info, "first interceptor must know the someUnaryServerInfo")
		ctx = context.WithValue(ctx, "first", 1)
		return handler(ctx, req)
	}
	second := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requireContextValue(t, ctx, "parent", "second interceptor must know the parent context value")
		requireContextValue(t, ctx, "first", "second interceptor must know the first context value")
		require.Equal(t, parentUnaryInfo, info, "second interceptor must know the someUnaryServerInfo")
		ctx = context.WithValue(ctx, "second", 1)
		return handler(ctx, req)
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		require.EqualValues(t, input, req, "handler must get the input")
		requireContextValue(t, ctx, "parent", "handler must know the parent context value")
		requireContextValue(t, ctx, "first", "handler must know the first context value")
		requireContextValue(t, ctx, "second", "handler must know the second context value")
		return output, nil
	}

	chain := ChainUnaryServer(first, second)
	out, _ := chain(parentContext, input, parentUnaryInfo, handler)
	require.EqualValues(t, output, out, "chain must return handler's output")
}

func TestChainStreamServer(t *testing.T) {
	someService := &struct{}{}
	recvMessage := "received"
	sentMessage := "sent"
	outputError := fmt.Errorf("some error")

	first := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		requireContextValue(t, stream.Context(), "parent", "first interceptor must know the parent context value")
		require.Equal(t, parentStreamInfo, info, "first interceptor must know the parentStreamInfo")
		require.Equal(t, someService, srv, "first interceptor must know someService")
		wrapped := WrapServerStream(stream)
		wrapped.WrappedContext = context.WithValue(stream.Context(), "first", 1)
		return handler(srv, wrapped)
	}
	second := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		requireContextValue(t, stream.Context(), "parent", "second interceptor must know the parent context value")
		requireContextValue(t, stream.Context(), "parent", "second interceptor must know the first context value")
		require.Equal(t, parentStreamInfo, info, "second interceptor must know the parentStreamInfo")
		require.Equal(t, someService, srv, "second interceptor must know someService")
		wrapped := WrapServerStream(stream)
		wrapped.WrappedContext = context.WithValue(stream.Context(), "second", 1)
		return handler(srv, wrapped)
	}
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		require.Equal(t, someService, srv, "handler must know someService")
		requireContextValue(t, stream.Context(), "parent", "handler must know the parent context value")
		requireContextValue(t, stream.Context(), "first", "handler must know the first context value")
		requireContextValue(t, stream.Context(), "second", "handler must know the second context value")
		require.NoError(t, stream.RecvMsg(recvMessage), "handler must have access to stream messages")
		require.NoError(t, stream.SendMsg(sentMessage), "handler must be able to send stream messages")
		return outputError
	}
	fakeStream := &fakeServerStream{ctx: parentContext, recvMessage: recvMessage}
	chain := ChainStreamServer(first, second)
	err := chain(someService, fakeStream, parentStreamInfo, handler)
	require.Equal(t, outputError, err, "chain must return handler's error")
	require.Equal(t, sentMessage, fakeStream.sentMessage, "handler's sent message must propagate to stream")
}

func requireContextValue(t *testing.T, ctx context.Context, key string, msg ...interface{}) {
	val := ctx.Value(key)
	require.NotNil(t, val, msg...)
	require.Equal(t, someValue, val, msg...)
}
