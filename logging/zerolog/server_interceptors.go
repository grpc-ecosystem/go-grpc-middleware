// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package grpc_zerolog

import (
	"context"
	"path"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

const (
	// SystemField is used in every log statement made through grpc_zerolog. Can be overwritten before any initialization code.
	SystemField = "system"

	// KindField describes the log field used to indicate whether this is a server or a client log statement.
	KindField = "span.kind"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zerolog.Logger to the context.
func UnaryServerInterceptor(logger *zerolog.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		newCtx := newLoggerForCall(ctx, logger, info.FullMethod, startTime)

		resp, err := handler(newCtx, req)

		if !o.shouldLog(info.FullMethod, err) {
			return resp, err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		durField, durVal := o.durationFunc(time.Since(startTime))
		fields := map[string]interface{}{
			"grpc.code": code.String(),
			durField:    durVal,
		}

		o.messageFunc(newCtx, "finished unary call with code "+code.String(), level, code, err, fields)
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds zerolog.Logger to the context.
func StreamServerInterceptor(logger *zerolog.Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		newCtx := newLoggerForCall(stream.Context(), logger, info.FullMethod, startTime)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		err := handler(srv, wrapped)

		if !o.shouldLog(info.FullMethod, err) {
			return err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		durField, durVal := o.durationFunc(time.Since(startTime))
		fields := map[string]interface{}{
			"grpc.code": code.String(),
			durField:    durVal,
		}

		o.messageFunc(newCtx, "finished streaming call with code "+code.String(), level, code, err, fields)
		return err
	}
}

func newLoggerForCall(ctx context.Context, logger *zerolog.Logger, fullMethodString string, start time.Time) context.Context {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)

	ctxContext := *ctxzerolog.Extract(ctx)
	ctxContext = ctxContext.Fields(
		map[string]interface{}{
			SystemField:       "grpc",
			KindField:         "server",
			"grpc.service":    service,
			"grpc.method":     method,
			"grpc.start_time": start.Format(time.RFC3339),
		})

	if d, ok := ctx.Deadline(); ok {
		ctxContext = ctxContext.Fields(
			map[string]interface{}{
				"grpc.request.deadline": d.Format(time.RFC3339),
			})
	}

	// Replace the logger's Context with our own new Context.
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return ctxContext
	})
	newContext := logger.With()

	return ctxzerolog.ToContext(ctx, &newContext)
}
