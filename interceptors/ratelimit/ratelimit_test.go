package ratelimit

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const errMsgFake = "fake error"

var ctxLimitKey = struct{}{}

type mockGRPCServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (m *mockGRPCServerStream) Context() context.Context {
	return m.ctx
}

type mockContextBasedLimiter struct{}

func (*mockContextBasedLimiter) Limit(ctx context.Context) bool {
	l, ok := ctx.Value(ctxLimitKey).(bool)
	return ok && l
}

func TestUnaryServerInterceptor_RateLimitPass(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxLimitKey, false)

	interceptor := UnaryServerInterceptor(limiter)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
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
	ctx := context.WithValue(context.Background(), ctxLimitKey, false)

	interceptor := StreamServerInterceptor(limiter)
	handler := func(srv interface{}, stream grpc.ServerStream) error {
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
	ctx := context.WithValue(context.Background(), ctxLimitKey, true)

	interceptor := UnaryServerInterceptor(limiter)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}
	resp, err := interceptor(ctx, nil, info, handler)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = FakeMethod is rejected by grpc_ratelimit middleware, please retry later.")
}

func TestStreamServerInterceptor_RateLimitFail(t *testing.T) {
	limiter := new(mockContextBasedLimiter)
	ctx := context.WithValue(context.Background(), ctxLimitKey, true)

	interceptor := StreamServerInterceptor(limiter)
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	info := &grpc.StreamServerInfo{
		FullMethod: "FakeMethod",
	}
	err := interceptor(nil, &mockGRPCServerStream{ctx: ctx}, info, handler)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = FakeMethod is rejected by grpc_ratelimit middleware, please retry later.")
}
