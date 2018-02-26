// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog

import (
	"path"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	// SystemField is used in every log statement made through grpc_glog. Can be overwritten before any initialization code.
	SystemField = "system"

	// KindField describes the log gield used to incicate whether this is a server or a client log statment.
	KindField = "span.kind"
)

// PayloadUnaryServerInterceptor returns a new unary server interceptors that adds ctx_glog.Entry to the context.
func UnaryServerInterceptor(entry *ctx_glog.Entry, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		newCtx := newLoggerForCall(ctx, entry, info.FullMethod, startTime)

		resp, err := handler(newCtx, req)

		if !o.shouldLog(info.FullMethod, err) {
			return resp, err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		durField, durVal := o.durationFunc(time.Since(startTime))
		fields := ctx_glog.Fields{
			"grpc.code": code.String(),
			durField:    durVal,
		}
		if err != nil {
			fields[ctx_glog.ErrorKey] = err
		}

		levelLogf(
			ctx_glog.Extract(newCtx).WithFields(fields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
			level,
			"finished unary call with code "+code.String())

		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds ctx_glog.Entry to the context.
func StreamServerInterceptor(entry *ctx_glog.Entry, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		newCtx := newLoggerForCall(stream.Context(), entry, info.FullMethod, startTime)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		err := handler(srv, wrapped)

		if !o.shouldLog(info.FullMethod, err) {
			return err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		durField, durVal := o.durationFunc(time.Since(startTime))
		fields := ctx_glog.Fields{
			"grpc.code": code.String(),
			durField:    durVal,
		}
		if err != nil {
			fields[ctx_glog.ErrorKey] = err
		}

		levelLogf(
			ctx_glog.Extract(newCtx).WithFields(fields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
			level,
			"finished streaming call with code "+code.String())

		return err
	}
}

func levelLogf(entry *ctx_glog.Entry, level ctx_glog.Severity, format string, args ...interface{}) {
	switch level {
	case ctx_glog.DebugLevel:
		entry.Debugf(format, args...)
	case ctx_glog.InfoLevel:
		entry.Infof(format, args...)
	case ctx_glog.WarningLevel:
		entry.Warningf(format, args...)
	case ctx_glog.ErrorLevel:
		entry.Errorf(format, args...)
	case ctx_glog.FatalLevel:
		entry.Fatalf(format, args...)
	}
}

func newLoggerForCall(ctx context.Context, entry *ctx_glog.Entry, fullMethodString string, start time.Time) context.Context {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	callLog := entry.WithFields(
		ctx_glog.Fields{
			SystemField:       "grpc",
			KindField:         "server",
			"grpc.service":    service,
			"grpc.method":     method,
			"grpc.start_time": start.Format(time.RFC3339),
		})

	if d, ok := ctx.Deadline(); ok {
		callLog = callLog.WithFields(
			ctx_glog.Fields{
				"grpc.request.deadline": d.Format(time.RFC3339),
			})
	}

	callLog = callLog.WithFields(ctx_glog.Extract(ctx).Data)
	return ctx_glog.ToContext(ctx, callLog)
}
