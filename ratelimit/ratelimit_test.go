package grpc_ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit/tokenbucket"
	"github.com/stretchr/testify/assert"
)

const errMsgFake = "fake error"

func TestEmptyUnaryServerInterceptor(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	var ctx context.Context
	var req interface{}
	var info *grpc.UnaryServerInfo
	req2, err := interceptor(ctx, req, info, handler)
	assert.Nil(t, req2)
	assert.EqualError(t, err, errMsgFake)
}

func TestRateLimitUnaryServerInterceptor(t *testing.T) {
	unaryRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 1, 1)
	interceptor := UnaryServerInterceptor(
		WithLimiter(unaryRateLimiter),
		WithMaxWaitDuration(1*time.Millisecond),
	)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	var ctx context.Context
	var req interface{}
	var info *grpc.UnaryServerInfo
	req2, err := interceptor(ctx, req, info, handler)
	assert.Nil(t, req2)
	assert.EqualError(t, err, errMsgFake)
}

func TestRateLimitStreamServerInterceptor(t *testing.T) {
	unaryRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 1, 1)
	interceptor := StreamServerInterceptor(
		WithLimiter(unaryRateLimiter),
		WithMaxWaitDuration(1*time.Millisecond),
	)
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	var srv interface{}
	var ss *mockServerStream
	var info *grpc.StreamServerInfo
	err := interceptor(srv, ss, info, handler)
	assert.EqualError(t, err, errMsgFake)
}

type mockServerStream struct{}

func (mss *mockServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (mss *mockServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (mss *mockServerStream) SetTrailer(metadata.MD) {}

func (mss *mockServerStream) Context() context.Context {
	return context.Background()
}

func (mss *mockServerStream) SendMsg(m interface{}) error {
	return nil
}

func (mss *mockServerStream) RecvMsg(m interface{}) error {
	return nil
}
