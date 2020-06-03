package ctxzerolog

import (
	"context"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/rs/zerolog"
)

type ctxLoggerMarker struct{}

type ctxLogger struct {
	logContext zerolog.Context
}

var (
	ctxLoggerKey = &ctxLoggerMarker{}
	nullContext  = zerolog.Nop().With()

	// NullContext is the NOP zerolog.Context we return when none has been stored before.
	NullContext = &nullContext
)

// AddFields adds zerolog fields to the Context.
func AddFields(ctx context.Context, fields map[string]interface{}) {
	l, ok := ctx.Value(ctxLoggerKey).(*ctxLogger)
	if !ok || l == nil {
		return
	}
	l.logContext = l.logContext.Fields(fields)
}

// Extract takes the call-scoped zerolog.Context from ctxzerolog middleware.
//
// If the ctxzerolog middleware wasn't used, a no-op `zerolog.Context` is returned. This makes it
// safe to use regardless.
func Extract(ctx context.Context) *zerolog.Context {
	if ctx == nil {
		return NullContext
	}
	l, ok := ctx.Value(ctxLoggerKey).(*ctxLogger)
	if !ok || l == nil {
		return NullContext
	}

	// Add grpc_ctxtags tags metadata until now.
	tags := grpc_ctxtags.Extract(ctx)
	values := tags.Values()
	fields := make(map[string]interface{}, len(values))
	for k, v := range values {
		fields[k] = v
	}

	// Add ctxLogger fields added until now.
	l.logContext = l.logContext.Fields(fields)

	return &l.logContext
}

// ToContext adds the zerolog.Contextd to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logContext *zerolog.Context) context.Context {
	if logContext == nil {
		return ctx
	}
	l := &ctxLogger{
		logContext: *logContext,
	}
	return context.WithValue(ctx, ctxLoggerKey, l)
}
