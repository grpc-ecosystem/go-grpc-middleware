package filter

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryServerMethods returns an interceptor that applies the provided interceptor only to incoming unary calls to the specified methods.
// The allowlist parameter specifies whether the provided list of methods is to be treated as an allowlist (true) or a denylist (false).
// If it is an allowlist the interceptor will be applied only to the methods in the list; if it is a denylist the interceptor will be applied only to methods not in the list.
// The methods must be specified using the full name (e.g. "/package.service/method").
func UnaryServerMethods(interceptor grpc.UnaryServerInterceptor, allowlist bool, methods ...string) grpc.UnaryServerInterceptor {
	if interceptor == nil {
		panic("nil interceptor")
	}
	m := newMatchlist(methods, allowlist)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if m.match(info.FullMethod) {
			return interceptor(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
}

/*
func UnaryServerMethodsInterceptor(interceptor grpc.UnaryServerInterceptor, allowlist bool, methods ...string) grpc.UnaryServerInterceptor {
	m := newMatchlist(methods, allowlist)

	return UnaryServerConditionInterceptor(interceptor, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) bool {
		return m.match(info.FullMethod)
	})
}

func UnaryServerConditionInterceptor(interceptor grpc.UnaryServerInterceptor, condition func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) bool) grpc.UnaryServerInterceptor {
	if interceptor == nil {
		panic("nil interceptor")
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if condition(ctx, req, info) {
			return interceptor(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
}
*/

// StreamServerMethods returns an interceptor that applies the provided interceptor only to incoming stream calls to the specified methods.
// The allowlist parameter specifies whether the provided list of methods is to be treated as an allowlist (true) or a denylist (false).
// If it is an allowlist the interceptor will be applied only to the methods in the list; if it is a denylist the interceptor will be applied only to methods not in the list.
// The methods must be specified using the full name (e.g. "/package.service/method").
func StreamServerMethods(interceptor grpc.StreamServerInterceptor, allowlist bool, methods ...string) grpc.StreamServerInterceptor {
	if interceptor == nil {
		panic("nil interceptor")
	}
	m := newMatchlist(methods, allowlist)

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if m.match(info.FullMethod) {
			return interceptor(srv, ss, info, handler)
		}
		return handler(srv, ss)
	}
}
