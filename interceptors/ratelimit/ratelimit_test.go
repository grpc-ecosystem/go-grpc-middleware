// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package ratelimit

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errMsgFake = "fake error"

type ctxKey string

var ctxKeyShouldLimit = ctxKey("shouldLimit")

type mockGRPCServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (m *mockGRPCServerStream) Context() context.Context {
	return m.ctx
}

type mockContextBasedLimiter struct{}

func (*mockContextBasedLimiter) Limit(ctx context.Context) error {
	shouldLimit, _ := ctx.Value(ctxKeyShouldLimit).(bool)

	if shouldLimit {
		return errors.New("rate limit exceeded")
	}

	return nil
}

func TestUnaryServerInterceptor_RateLimitPass(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, false)

	interceptor := UnaryServerInterceptor(limiter)
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, errors.New(errMsgFake)
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}
	resp, err := interceptor(ctx, nil, info, handler)
	assert.Nil(t, resp)
	assert.EqualError(t, err, errMsgFake)
}

func TestStreamServerInterceptor_RateLimitPass(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, false)

	interceptor := StreamServerInterceptor(limiter)
	handler := func(srv any, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	info := &grpc.StreamServerInfo{
		FullMethod: "FakeMethod",
	}
	err := interceptor(nil, &mockGRPCServerStream{ctx: ctx}, info, handler)
	assert.EqualError(t, err, errMsgFake)
}

func TestUnaryServerInterceptor_RateLimitFail(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, true)

	interceptor := UnaryServerInterceptor(limiter)
	called := false
	handler := func(ctx context.Context, req any) (any, error) {
		called = true
		return nil, errors.New(errMsgFake)
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}
	resp, err := interceptor(ctx, nil, info, handler)
	expErr := status.Errorf(
		codes.ResourceExhausted,
		"%s is rejected by grpc_ratelimit middleware, please retry later. %s",
		info.FullMethod,
		"rate limit exceeded",
	)
	assert.Nil(t, resp)
	assert.EqualError(t, err, expErr.Error())
	assert.False(t, called)
}

func TestStreamServerInterceptor_RateLimitFail(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, true)

	interceptor := StreamServerInterceptor(limiter)
	called := false
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		called = true
		return errors.New(errMsgFake)
	}
	info := &grpc.StreamServerInfo{
		FullMethod: "FakeMethod",
	}
	err := interceptor(nil, &mockGRPCServerStream{ctx: ctx}, info, handler)
	expErr := status.Errorf(
		codes.ResourceExhausted,
		"%s is rejected by grpc_ratelimit middleware, please retry later. %s",
		info.FullMethod,
		"rate limit exceeded",
	)

	assert.EqualError(t, err, expErr.Error())
	assert.False(t, called)
}

func TestUnaryClientInterceptor_RateLimitPass(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, false)

	interceptor := UnaryClientInterceptor(limiter)
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return errors.New(errMsgFake)
	}
	err := interceptor(ctx, "FakeMethod", nil, nil, nil, invoker)
	assert.EqualError(t, err, errMsgFake)
}

func TestStreamClientInterceptor_RateLimitPass(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, false)

	interceptor := StreamClientInterceptor(limiter)
	invoker := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, errors.New(errMsgFake)
	}
	_, err := interceptor(ctx, nil, nil, "FakeMethod", invoker)
	assert.EqualError(t, err, errMsgFake)
}

func TestUnaryClientInterceptor_RateLimitFail(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, true)

	interceptor := UnaryClientInterceptor(limiter)
	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		called = true
		return errors.New(errMsgFake)
	}
	err := interceptor(ctx, "FakeMethod", nil, nil, nil, invoker)
	expErr := status.Errorf(
		codes.ResourceExhausted,
		"%s is rejected by grpc_ratelimit middleware, please retry later. %s",
		"FakeMethod",
		"rate limit exceeded",
	)
	assert.EqualError(t, err, expErr.Error())
	assert.False(t, called)
}

func TestStreamClientInterceptor_RateLimitFail(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxKeyShouldLimit, true)

	interceptor := StreamClientInterceptor(limiter)
	called := false
	invoker := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		called = true
		return nil, errors.New(errMsgFake)
	}
	_, err := interceptor(ctx, nil, nil, "FakeMethod", invoker)
	expErr := status.Errorf(
		codes.ResourceExhausted,
		"%s is rejected by grpc_ratelimit middleware, please retry later. %s",
		"FakeMethod",
		"rate limit exceeded",
	)
	assert.EqualError(t, err, expErr.Error())
	assert.False(t, called)
}
