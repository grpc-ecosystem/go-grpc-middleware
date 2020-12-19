package grpc_auth

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// AuthzFunc is the pluggable function that performs authorization.
//
// The passed in `Context` will contain the gRPC metadata.MD object (for header-based authorization).
// `authorization bearer <token>` for example
//
// The returned context will be propagated to handlers, allowing user changes to `Context`. However,
// please make sure that the `Context` returned is a child `Context` of the one passed in.
//
// If error is returned, its `grpc.Code()` will be returned to the user as well as the verbatim message.
// Please make sure you use `codes.Unauthenticated` (lacking auth) and `codes.PermissionDenied`
// (authed, but lacking perms) appropriately.
type AuthzFunc func(ctx context.Context, fullMethodName string) (context.Context, error)

// ServiceAuthzFuncOverride allows a given gRPC service implementation to override the global `AuthzFunc`.
//
// If a service implements the AuthzFuncOverride method, it takes precedence over the `AuthzFunc` method,
// and will be called instead of AuthzFunc for all method invocations within that service.
type ServiceAuthzFuncOverride interface {
	AuthzFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error)
}

// UnaryServerInterceptorAuthz returns a new unary server interceptors that performs per-request authorization.
func UnaryServerInterceptorAuthz(authzFunc AuthzFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var newCtx context.Context
		var err error
		if overrideSrv, ok := info.Server.(ServiceAuthzFuncOverride); ok {
			newCtx, err = overrideSrv.AuthzFuncOverride(ctx, info.FullMethod)
		} else {
			newCtx, err = authzFunc(ctx, info.FullMethod)
		}
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// StreamServerInterceptorAuthz returns a new unary server interceptors that performs per-request authorization.
func StreamServerInterceptorAuthz(authzFunc AuthzFunc) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var newCtx context.Context
		var err error
		if overrideSrv, ok := srv.(ServiceAuthzFuncOverride); ok {
			newCtx, err = overrideSrv.AuthzFuncOverride(stream.Context(), info.FullMethod)
		} else {
			newCtx, err = authzFunc(stream.Context(), info.FullMethod)
		}
		if err != nil {
			return err
		}
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}
