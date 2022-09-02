// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package selector

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
)

var blockList = []string{"/auth.v1beta1.AuthService/Login"}

const errMsgFake = "fake error"

var ctxKey = struct{}{}

// allow After the method is matched, the interceptor is run
func allow(methods []string) MatchFunc {
	return func(ctx context.Context, fullMethod string) bool {
		for _, s := range methods {
			if s == fullMethod {
				return true
			}
		}
		return false
	}
}

// Block the interceptor will not run after the method matches
func block(methods []string) MatchFunc {
	allow := allow(methods)
	return func(ctx context.Context, fullMethod string) bool {
		return !allow(ctx, fullMethod)
	}
}

type mockGRPCServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (m *mockGRPCServerStream) Context() context.Context {
	return m.ctx
}

func TestUnaryServerInterceptor(t *testing.T) {
	ctx := context.Background()
	interceptor := UnaryServerInterceptor(auth.UnaryServerInterceptor(
		func(ctx context.Context) (context.Context, error) {
			newCtx := context.WithValue(ctx, ctxKey, true)
			return newCtx, nil
		},
	), block(blockList))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		val := ctx.Value(ctxKey)
		if b, ok := val.(bool); ok && b {
			return "good", nil
		}
		return nil, errors.New(errMsgFake)
	}

	t.Run("nextStep", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
		resp, err := interceptor(ctx, nil, info, handler)
		assert.Nil(t, err)
		assert.Equal(t, resp, "good")
	})

	t.Run("skipped", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "/auth.v1beta1.AuthService/Login",
		}
		resp, err := interceptor(ctx, nil, info, handler)
		assert.Nil(t, resp)
		assert.EqualError(t, err, errMsgFake)
	})
}

func TestStreamServerInterceptor(t *testing.T) {
	ctx := context.Background()
	interceptor := StreamServerInterceptor(auth.StreamServerInterceptor(
		func(ctx context.Context) (context.Context, error) {
			newCtx := context.WithValue(ctx, ctxKey, true)
			return newCtx, nil
		},
	), block(blockList))

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		ctx := stream.Context()
		val := ctx.Value(ctxKey)
		if b, ok := val.(bool); ok && b {
			return nil
		}
		return errors.New(errMsgFake)
	}

	t.Run("nextStep", func(t *testing.T) {
		info := &grpc.StreamServerInfo{
			FullMethod: "FakeMethod",
		}
		err := interceptor(nil, &mockGRPCServerStream{ctx: ctx}, info, handler)
		assert.Nil(t, err)
	})

	t.Run("skipped", func(t *testing.T) {
		info := &grpc.StreamServerInfo{
			FullMethod: "/auth.v1beta1.AuthService/Login",
		}
		err := interceptor(nil, &mockGRPCServerStream{ctx: ctx}, info, handler)
		assert.EqualError(t, err, errMsgFake)
	})
}

func TestAllow(t *testing.T) {
	type args struct {
		methods []string
	}
	tests := []struct {
		name   string
		args   args
		method string
		want   bool
	}{
		{
			name: "false",
			args: args{
				methods: []string{"/auth.v1beta1.AuthService/Login"},
			},
			method: "/testing.testpb.v1.TestService/PingList",
			want:   false,
		},
		{
			name: "true",
			args: args{
				methods: []string{"/auth.v1beta1.AuthService/Login"},
			},
			method: "/auth.v1beta1.AuthService/Login",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allow := allow(tt.args.methods)
			want := allow(context.Background(), tt.method)
			assert.Equalf(t, tt.want, want, "Allow(%v)(ctx, %v)", tt.args.methods, tt.method)
		})
	}
}
