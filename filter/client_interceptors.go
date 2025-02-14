package filter

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryClientMethods returns an interceptor that applies the provided interceptor only to outgoing unary calls to the specified methods.
// The allowlist parameter specifies whether the provided list of methods is to be treated as an allowlist (true) or a denylist (false).
// If it is an allowlist the interceptor will be applied only to the methods in the list; if it is a denylist the interceptor will be applied only to methods not in the list.
// The methods must be specified using the full name (e.g. "/package.service/method").
func UnaryClientMethods(interceptor grpc.UnaryClientInterceptor, allowlist bool, methods ...string) grpc.UnaryClientInterceptor {
	if interceptor == nil {
		panic("nil interceptor")
	}
	m := newMatchlist(methods, allowlist)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if m.match(method) {
			return interceptor(ctx, method, req, reply, cc, invoker, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientMethods returns an interceptor that applies the provided interceptor only to outgoing unary calls to the specified methods.
// The allowlist parameter specifies whether the provided list of methods is to be treated as an allowlist (true) or a denylist (false).
// If it is an allowlist the interceptor will be applied only to the methods in the list; if it is a denylist the interceptor will be applied only to methods not in the list.
// The methods must be specified using the full name (e.g. "/package.service/method").
func StreamClientMethods(interceptor grpc.StreamClientInterceptor, allowlist bool, methods ...string) grpc.StreamClientInterceptor {
	if interceptor == nil {
		panic("nil interceptor")
	}
	m := newMatchlist(methods, allowlist)

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if m.match(method) {
			return interceptor(ctx, desc, cc, method, streamer, opts...)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
