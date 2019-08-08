// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_kit

import (
	"path"
	"time"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
// ClientField is used in every client-side log statement made through grpc_zap. Can be overwritten before initialization.
//ClientField = zap.String("span.kind", "client")
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(logger log.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		logFinalClientLine(o, log.With(logger, fields...), startTime, err, "finished client unary call")
		return err
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(logger log.Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		logFinalClientLine(o, log.With(logger, fields...), startTime, err, "finished client streaming call")
		return clientStream, err
	}
}

func logFinalClientLine(o *options, logger log.Logger, startTime time.Time, err error, msg string) {
	code := o.codeFunc(err)
	logger = o.levelFunc(code, logger)
	args := []interface{}{"msg", msg, "error", err, "grpc.code", code.String()}
	args = append(args, o.durationFunc(time.Now().Sub(startTime))...)
	_ = logger.Log(args...)
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
