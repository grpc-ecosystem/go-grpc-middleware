package skip

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

type Filter func(ctx context.Context, gRPCType interceptors.GRPCType, service string, method string) bool

// UnaryServerInterceptor returns a new unary server interceptor that determines whether to skip the input interceptor.
func UnaryServerInterceptor(in grpc.UnaryServerInterceptor, filter Filter) grpc.UnaryServerInterceptor {
	if filter == nil {
		return in
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		service, method := interceptors.SplitMethodName(info.FullMethod)
		if filter(ctx, interceptors.Unary, service, method) {
			// Skip interceptor
			return handler(ctx, req)
		}
		return in(ctx, req, info, handler)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that determines whether to skip the input interceptor.
func StreamServerInterceptor(in grpc.StreamServerInterceptor, filter Filter) grpc.StreamServerInterceptor {
	if filter == nil {
		return in
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		service, method := interceptors.SplitMethodName(info.FullMethod)
		if filter(ss.Context(), interceptors.StreamRPCType(info), service, method) {
			// Skip interceptor
			return handler(srv, ss)
		}
		return in(srv, ss, info, handler)
	}
}
