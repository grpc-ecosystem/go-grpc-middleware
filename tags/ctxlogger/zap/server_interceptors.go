package ctxlogger_zap

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newLoggerForCall(ctx, logger)
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds zap.Logger to the context.
func StreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newLoggerForCall(stream.Context(), logger)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

func newLoggerForCall(ctx context.Context, logger *zap.Logger) context.Context {
	return ToContext(ctx, logger)
}
