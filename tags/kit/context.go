package ctx_kit

import (
	"github.com/go-kit/kit/log"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/kit/ctxkit"
	"golang.org/x/net/context"
)

// AddFields adds logrus fields to the logger.
// Deprecated: should use the ctxlogrus.Extract instead
func AddFields(ctx context.Context, fields []interface{}) {
	ctxkit.AddFields(ctx, fields...)
}

// Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.
// Deprecated: should use the ctxlogrus.Extract instead
func Extract(ctx context.Context) log.Logger {
	return ctxkit.Extract(ctx)
}

// ToContext adds the logrus.Entry to the context for extraction later.
// Depricated: should use ctxlogrus.ToContext instead
func ToContext(ctx context.Context, logger log.Logger) context.Context {
	return ctxkit.ToContext(ctx, logger)
}
