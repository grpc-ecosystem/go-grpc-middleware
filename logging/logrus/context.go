package grpc_logrus

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/ctxlogger/logrus"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// AddFields adds logrus fields to the logger.
// Deprecated: should use the ctxlogger_logrus.Extract instead
func AddFields(ctx context.Context, fields logrus.Fields) {
	ctxlogger_logrus.AddFields(ctx, fields)
}

// Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.
// Deprecated: should use the ctxlogger_logrus.Extract instead
func Extract(ctx context.Context) *logrus.Entry {
	return ctxlogger_logrus.Extract(ctx)
}
