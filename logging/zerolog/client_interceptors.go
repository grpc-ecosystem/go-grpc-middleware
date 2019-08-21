// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zerolog

import (
	"context"
	"fmt"
	"github.com/Ahmet-Kaplan/go-grpc-middleware/logging/zerolog/ctxzr"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"path"
	"time"
)

var (
// ClientField is used in every client-side log statement made through grpc_zap. Can be overwritten before initialization.
//ClientField = zap.String("span.kind", "client")
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(logger *zerolog.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		logFinalClientLine(o, &ctxzr.CtxLogger{Logger: logger, Fields: fields}, startTime, err, "finished client unary call")
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
		logFinalClientLine(o, &ctxzr.CtxLogger{Logger: logger, Fields: fields}, startTime, err, "finished client streaming call")
		return clientStream, err
	}
}

func logFinalClientLine(o *options, logger *ctxzr.CtxLogger, startTime time.Time, err error, msg string) {
	code := o.codeFunc(err)
	var level = o.levelFunc(code)
	args := []interface{}{"msg", msg, "error", err, "grpc.code", code.String()}
	args = append(args, o.durationFunc(time.Now().Sub(startTime))...)
	//loggerEvent.Str("msg",msg).Err(err).Str("grpc.code",code.String()).Str("time",time.Now().Sub(startTime).String())
	logger.Logger.WithLevel(level).Msg(fmt.Sprint(args...))

}

func newClientLoggerFields(ctx context.Context, fullMethodString string) []interface{} {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []interface{}{
		"system", "grpc",
		"span.kind", "client",
		"grpc.service", service,
		"grpc.method", method,
	}
}
