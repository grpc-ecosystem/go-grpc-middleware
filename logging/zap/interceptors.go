// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zap

import (
	"path"
	"time"

	"github.com/mwitkow/go-grpc-middleware/logging"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "grpc")
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryServerInterceptor(logger *zap.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newLoggerForCall(ctx, logger, info.FullMethod)
		keys, values := o.fieldExtractorFunc(info.FullMethod, req)
		if keys != nil && values != nil {
			grpc_logging.ExtractMetadata(newCtx).AddFieldsFromMiddleware(keys, values)
		}

		startTime := time.Now()
		resp, err := handler(newCtx, req)
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		// re-extract logger from newCtx, as it may have extra fields that changed in the holder.

		Extract(newCtx).Check(level, "finished unary call").Write(
			zap.Error(err),
			zap.String("grpc_code", code.String()),
			zap.Duration("grpc_time_ns", time.Now().Sub(startTime)),
		)
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds zap.Logger to the context.
func StreamServerInterceptor(logger *zap.Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateOptions(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newLoggerForCall(stream.Context(), logger, info.FullMethod)
		wrapped := &wrappedStream{stream, info, o, newCtx}

		startTime := time.Now()
		err := handler(srv, wrapped)
		code := o.codeFunc(err)
		level := o.levelFunc(code)

		// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
		Extract(newCtx).Check(level, "finished streaming call").Write(
			zap.Error(err),
			zap.String("grpc_code", code.String()),
			zap.Duration("grpc_time_ns", time.Now().Sub(startTime)),
		)
		return err
	}
}

// wrappedStream is a thin wrapper around grpc.ServerStream that allows modifying context and extracts log fields from the initial message.
type wrappedStream struct {
	grpc.ServerStream
	info *grpc.StreamServerInfo
	opts *options
	// WrappedContext is the wrapper's own Context. You can assign it.
	WrappedContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *wrappedStream) Context() context.Context {
	return w.WrappedContext
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	// We only do log fields extraction on the single-reqest of a server-side stream.
	if !w.info.IsClientStream {
		keys, values := w.opts.fieldExtractorFunc(w.info.FullMethod, m)
		if keys != nil && values != nil {
			grpc_logging.ExtractMetadata(w.Context()).AddFieldsFromMiddleware(keys, values)
		}
	}
	return err
}

func newLoggerForCall(ctx context.Context, logger *zap.Logger, fullMethodString string) context.Context {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	callLog := logger.With(
		SystemField,
		zap.String("grpc_service", service),
		zap.String("grpc_method", method))
	return toContext(ctx, callLog)
}
