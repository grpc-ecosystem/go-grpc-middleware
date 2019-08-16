package ctx_zerolog

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzr"
)

// AddFields adds logrus fields to the logger.
// Deprecated: should use the ctxlogrus.Extract instead
func AddFields(ctx context.Context, fields []interface{}) {
	ctxzr.AddFields(ctx, fields...)
}

// Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.
// Deprecated: should use the ctxlogrus.Extract instead
func Extract(ctx context.Context) *ctxzr.CtxLogger {
	return ctxzr.Extract(ctx)
}

// ToContext adds the logrus.Entry to the context for extraction later.
// Depricated: should use ctxlogrus.ToContext instead
func ToContext(ctx context.Context, logger ctxzr.CtxLogger) context.Context {
	return ctxzr.ToContext(ctx, &logger)
}
