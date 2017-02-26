// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zap

import (
	"github.com/mwitkow/go-grpc-middleware/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
)

var (
	nullLogger = zap.NewNop()
)

type holder struct {
	*zap.Logger
}

// AddFieldsFromMiddleware implements grpc_logging.Metadata on this holder.
func (h *holder) AddFieldsFromMiddleware(keys []string, values []interface{}) {
	if len(keys) != len(values) {
		panic("AddFieldsFromMiddleware length of keys doesn't match length of values")
	}
	fields := []zapcore.Field{}
	for i := range keys {
		fields = append(fields, zap.Any(keys[i], values[i]))
	}
	h.Logger = h.Logger.With(fields...)
}

// Extract takes the call-scoped Logger from grpc_zap middleware.
//
// If the grpc_zap middleware wasn't used, a null `zap.Logger` is returned. This makes it safe to
// use regardless.
func Extract(ctx context.Context) *zap.Logger {
	h, ok := ctx.Value(grpc_logging.InternalContextMarker).(*holder)
	if !ok {
		return nullLogger
	}
	return h.Logger
}

// AddFields adds zap.Fields to *all* usages of the logger, both upstream (to handler) and downstream.
//
// This call *is not* concurrency safe. It should only be used in the request goroutine: in other
// interceptors or directly in the handler.
func AddFields(ctx context.Context, fields ...zapcore.Field) {
	logger := Extract(ctx)
	holder, ok := ctx.Value(grpc_logging.InternalContextMarker).(*holder)
	if !ok {
		return
	}
	holder.Logger = logger.With(fields...)
}

func toContext(ctx context.Context, logger *zap.Logger) context.Context {
	h, ok := ctx.Value(grpc_logging.InternalContextMarker).(*holder)
	if !ok {
		return context.WithValue(ctx, grpc_logging.InternalContextMarker, &holder{logger})
	}
	h.Logger = logger
	return ctx
}
