package ctxlogger_logrus

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// PayloadUnaryServerInterceptor returns a new unary server interceptors that adds logrus.Entry to the context.
func UnaryServerInterceptor(entry *logrus.Entry) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newLoggerForCall(ctx, entry)
		resp, err := handler(newCtx, req)
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds logrus.Entry to the context.
func StreamServerInterceptor(entry *logrus.Entry) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newLoggerForCall(stream.Context(), entry)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		err := handler(srv, wrapped)
		return err
	}
}

func newLoggerForCall(ctx context.Context, entry *logrus.Entry) context.Context {
	return ToContext(ctx, entry)
}
