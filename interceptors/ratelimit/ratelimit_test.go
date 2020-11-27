package ratelimit

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	errMsgFake = "fake error"
)

var (
	ctxLimitKey = struct{}{}
)

type mockGRPCServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (m *mockGRPCServerStream) Context() context.Context {
	return m.ctx
}

type mockPassLimiter struct{}

func (*mockPassLimiter) Limit(_ context.Context) bool {
	return false
}

type mockFailLimiter struct{}

func (*mockFailLimiter) Limit(_ context.Context) bool {
	return true
}

type mockContextBasedLimiter struct{}

func (*mockContextBasedLimiter) Limit(ctx context.Context) bool {
	l, ok := ctx.Value(ctxLimitKey).(bool)
	return ok && l
}

type testcase struct {
	limiter Limiter
	ctx     context.Context
}

func TestUnaryServerInterceptor_RateLimitPass(t *testing.T) {
	for _, tc := range []testcase{
		{
			limiter: new(mockPassLimiter),
		}, {
			limiter: new(mockContextBasedLimiter),
			ctx:     context.WithValue(context.Background(), ctxLimitKey, false),
		},
	} {
		interceptor := UnaryServerInterceptor(tc.limiter)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, errors.New(errMsgFake)
		}
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
		resp, err := interceptor(tc.ctx, nil, info, handler)
		assert.Nil(t, resp)
		assert.EqualError(t, err, errMsgFake)
	}
}

func TestStreamServerInterceptor_RateLimitPass(t *testing.T) {
	for _, tc := range []testcase{
		{
			limiter: new(mockPassLimiter),
		}, {
			limiter: new(mockContextBasedLimiter),
			ctx:     context.WithValue(context.Background(), ctxLimitKey, false),
		},
	} {
		interceptor := StreamServerInterceptor(tc.limiter)
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return errors.New(errMsgFake)
		}
		info := &grpc.StreamServerInfo{
			FullMethod: "FakeMethod",
		}
		err := interceptor(nil, &mockGRPCServerStream{ctx: tc.ctx}, info, handler)
		assert.EqualError(t, err, errMsgFake)
	}
}

func TestUnaryServerInterceptor_RateLimitFail(t *testing.T) {
	for _, tc := range []testcase{
		{
			limiter: new(mockFailLimiter),
		}, {
			limiter: new(mockContextBasedLimiter),
			ctx:     context.WithValue(context.Background(), ctxLimitKey, true),
		},
	} {
		interceptor := UnaryServerInterceptor(tc.limiter)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, errors.New(errMsgFake)
		}
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
		resp, err := interceptor(tc.ctx, nil, info, handler)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = FakeMethod is rejected by grpc_ratelimit middleware, please retry later.")
	}
}

func TestStreamServerInterceptor_RateLimitFail(t *testing.T) {
	for _, tc := range []testcase{
		{
			limiter: new(mockFailLimiter),
		}, {
			limiter: new(mockContextBasedLimiter),
			ctx:     context.WithValue(context.Background(), ctxLimitKey, true),
		},
	} {
		interceptor := StreamServerInterceptor(tc.limiter)
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return errors.New(errMsgFake)
		}
		info := &grpc.StreamServerInfo{
			FullMethod: "FakeMethod",
		}
		err := interceptor(nil, &mockGRPCServerStream{ctx: tc.ctx}, info, handler)
		assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = FakeMethod is rejected by grpc_ratelimit middleware, please retry later.")
	}
}
