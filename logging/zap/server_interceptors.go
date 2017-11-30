// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package grpc_zap

import (
	"path"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "grpc")

	// ServerField is used in every server-side log statment made through grpc_zap.Can be overwritten before initialization.
	ServerField = zap.String("span.kind", "server")
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryServerInterceptor(logger *zap.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newLoggerForCall(ctx, logger, info.FullMethod)
		startTime := time.Now()
		resp, err := handler(newCtx, req)
		if !o.shouldLog(info.FullMethod, err) {
			return resp, err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)

		// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
		ctx_zap.Extract(newCtx).Check(level, "finished unary call").Write(
			zap.Error(err),
			zap.String("grpc.code", code.String()),
			o.durationFunc(time.Now().Sub(startTime)),
		)

		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds zap.Logger to the context.
func StreamServerInterceptor(logger *zap.Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newLoggerForCall(stream.Context(), logger, info.FullMethod)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		startTime := time.Now()
		err := handler(srv, wrapped)
		if !o.shouldLog(info.FullMethod, err) {
			return err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)

		// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
		ctx_zap.Extract(newCtx).Check(level, "finished streaming call").Write(
			zap.Error(err),
			zap.String("grpc.code", code.String()),
			o.durationFunc(time.Now().Sub(startTime)),
		)

		return err
	}
}

func serverCallFields(fullMethodString string) []zapcore.Field {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []zapcore.Field{
		SystemField,
		ServerField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}

func newLoggerForCall(ctx context.Context, logger *zap.Logger, fullMethodString string) context.Context {
	callLog := logger.With(append(ctx_zap.TagsToFields(ctx), serverCallFields(fullMethodString)...)...)
	return ctx_zap.ToContext(ctx, callLog)
}
