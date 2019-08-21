package ctxzr

import (
	"context"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/rs/zerolog"
)

type ctxMarker struct{}

type CtxLogger struct {
	Logger *zerolog.Logger
	Fields []interface{}
}

var (
	ctxMarkerKey = &ctxMarker{}
	nullLogger   = &zerolog.Logger{}
)

// AddFields adds fields to the logger.
func AddFields(ctx context.Context, fields ...interface{}) {
	l, ok := ctx.Value(ctxMarkerKey).(*CtxLogger)
	if !ok || l == nil {
		return
	}
	l.Fields = append(l.Fields, fields...)
}

// Extract takes the call-scoped Logger from grpc_kit middleware.
//
// It always returns a Logger that has all the grpc_ctxtags updated.
func Extract(ctx context.Context) *CtxLogger {
	l, ok := ctx.Value(ctxMarkerKey).(*CtxLogger)
	if !ok || l == nil {
		return &CtxLogger{Logger: nullLogger, Fields: nil}
	}
	// Add grpc_ctxtags tags metadata until now.
	fields := TagsToFields(ctx)
	// Addfields added until now.
	fields = append(fields, l.Fields...)
	return &CtxLogger{Logger: l.Logger, Fields: fields}
}

// TagsToFields transforms the Tags on the supplied context into zap fields.
func TagsToFields(ctx context.Context) []interface{} {
	fields := []interface{}{}
	tags := grpc_ctxtags.Extract(ctx)
	for k, v := range tags.Values() {
		fields = append(fields, k, v)
	}
	return fields
}

// ToContext adds the zap.Logger to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logger *CtxLogger) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, logger)
}
