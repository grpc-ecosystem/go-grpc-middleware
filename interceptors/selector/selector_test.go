// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package selector

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

// allow matches only given methods.
func allow(methods []string) Matcher {
	return MatchFunc(func(ctx context.Context, c interceptors.CallMeta) bool {
		for _, s := range methods {
			if s == c.FullMethod() {
				return true
			}
		}
		return false
	})
}

type mockGRPCServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (m *mockGRPCServerStream) Context() context.Context {
	return m.ctx
}

const svcMethod = "/v1beta1.SomeService/NeedsAuth"

func TestUnaryServerInterceptor(t *testing.T) {
	interceptor := UnaryServerInterceptor(
		func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (any, error) {
			return nil, errors.New("always error")
		}, allow([]string{svcMethod}),
	)

	handler := func(ctx context.Context, req any) (any, error) {
		return "good", nil
	}

	t.Run("not-selected", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, err)
		assert.Equal(t, resp, "good")
	})

	t.Run("selected", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: svcMethod,
		}
		resp, err := interceptor(context.Background(), nil, info, handler)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "always error")
	})
}

func TestStreamServerInterceptor(t *testing.T) {
	interceptor := StreamServerInterceptor(
		func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return errors.New("always error")
		},
		allow([]string{svcMethod}),
	)

	handler := func(srv any, stream grpc.ServerStream) error {
		return nil
	}

	t.Run("not-selected", func(t *testing.T) {
		info := &grpc.StreamServerInfo{
			FullMethod: "FakeMethod",
		}

		err := interceptor(nil, &mockGRPCServerStream{ctx: context.Background()}, info, handler)
		assert.Nil(t, err)
	})

	t.Run("slected", func(t *testing.T) {
		info := &grpc.StreamServerInfo{
			FullMethod: svcMethod,
		}
		err := interceptor(nil, &mockGRPCServerStream{ctx: context.Background()}, info, handler)
		assert.EqualError(t, err, "always error")
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
			want := allow.Match(context.Background(), interceptors.NewServerCallMeta(tt.method, nil, nil))
			assert.Equalf(t, tt.want, want, "Allow(%v)(ctx, %v)", tt.args.methods, tt.method)
		})
	}
}
