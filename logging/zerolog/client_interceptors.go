// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package zerolog

import (
	"context"
	"path"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(logger *zerolog.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)

		ctxLogger := logger.With().Fields(fields)
		newCtx := ctxzerolog.ToContext(ctx, &ctxLogger)

		logFinalClientLine(newCtx, o, startTime, err, "finished client unary call")
		return err
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(logger *zerolog.Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		ctxLogger := logger.With().Fields(fields)
		newCtx := ctxzerolog.ToContext(ctx, &ctxLogger)

		logFinalClientLine(newCtx, o, startTime, err, "finished client streaming call")
		return clientStream, err
	}
}

func logFinalClientLine(ctx context.Context, o *options, startTime time.Time, err error, msg string) {
	code := o.codeFunc(err)
	level := o.levelFunc(code)
	durField, durVal := o.durationFunc(time.Since(startTime))
	fields := map[string]interface{}{
		"grpc.code": code.String(),
		durField:    durVal,
	}
	o.messageFunc(ctx, msg, level, code, err, fields)
}

func newClientLoggerFields(_ context.Context, fullMethodString string) map[string]interface{} {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return map[string]interface{}{
		SystemField:    "grpc",
		KindField:      "client",
		"grpc.service": service,
		"grpc.method":  method,
	}
}
