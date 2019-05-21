package ratelimit

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
)

const errMsgFake = "fake error"

func TestUnaryServerInterceptor_NoLimit(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	req, err := interceptor(nil, nil, nil, handler)
	assert.Nil(t, req)
	assert.EqualError(t, err, errMsgFake)
}

type mockPassLimiter struct{}

func (*mockPassLimiter) Limit() bool {
	return false
}

func TestUnaryServerInterceptor_RateLimitPass(t *testing.T) {
	unaryRateLimiter := &mockPassLimiter{}
	interceptor := UnaryServerInterceptor(
		WithRateLimiter(unaryRateLimiter),
	)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}
	req, err := interceptor(nil, nil, info, handler)
	assert.Nil(t, req)
	assert.EqualError(t, err, errMsgFake)
}

type mockFailLimiter struct{}

func (*mockFailLimiter) Limit() bool {
	return true
}

func TestUnaryServerInterceptor_RateLimitFail(t *testing.T) {
	unaryRateLimiter := &mockFailLimiter{}
	interceptor := UnaryServerInterceptor(
		WithRateLimiter(unaryRateLimiter),
	)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}
	req, err := interceptor(nil, nil, info, handler)
	assert.Nil(t, req)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = FakeMethod is rejected by grpc_ratelimit middleare, please retry later.")
}

func TestStreamServerInterceptor_NoLimit(t *testing.T) {
	interceptor := StreamServerInterceptor()
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	err := interceptor(nil, nil, nil, handler)
	assert.EqualError(t, err, errMsgFake)
}

func TestStreamServerInterceptor_RateLimitPass(t *testing.T) {
	streamRateLimiter := &mockPassLimiter{}
	interceptor := StreamServerInterceptor(
		WithRateLimiter(streamRateLimiter),
	)
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	info := &grpc.StreamServerInfo{
		FullMethod: "FakeMethod",
	}
	err := interceptor(nil, nil, info, handler)
	assert.EqualError(t, err, errMsgFake)
}

func TestStreamServerInterceptor_RateLimitFail(t *testing.T) {
	streamRateLimiter := &mockFailLimiter{}
	interceptor := StreamServerInterceptor(
		WithRateLimiter(streamRateLimiter),
	)
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	info := &grpc.StreamServerInfo{
		FullMethod: "FakeMethod",
	}
	err := interceptor(nil, nil, info, handler)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = FakeMethod is rejected by grpc_ratelimit middleare, please retry later.")
}
