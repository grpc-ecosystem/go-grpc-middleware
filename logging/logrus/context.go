// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logrus

import (
	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-grpc-middleware/logging"
	"golang.org/x/net/context"
)

type holder struct {
	*logrus.Entry
}

// AddFieldsFromMiddleware implements grpc_logging.Metadata on this holder.
func (h *holder) AddFieldsFromMiddleware(keys []string, values []interface{}) {
	if len(keys) != len(values) {
		panic("AddFieldsFromMiddleware length of keys doesn't match length of values")
	}
	fields := logrus.Fields{}
	for i := range keys {
		fields[keys[i]] = values[i]
	}
	h.Entry = h.Entry.WithFields(fields)
}

// Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.
//
// If the grpc_logrus middleware wasn't used, a no-op `logrus.Entry` is returned. This makes it safe to
// use regardless.
func Extract(ctx context.Context) *logrus.Entry {
	h, ok := ctx.Value(grpc_logging.InternalContextMarker).(*holder)
	if !ok {
		return logrus.NewEntry(nullLogger)
	}
	return h.Entry
}

// AddFields adds logrus.Fields to *all* usages of the logger, both upstream (to handler) and downstream (to interceptor).
//
// This call *is not* concurrency safe. It should only be used in the request goroutine: in other
// interceptors or directly in the handler.
func AddFields(ctx context.Context, fields logrus.Fields) {
	logger := Extract(ctx)
	holder, ok := ctx.Value(grpc_logging.InternalContextMarker).(*holder)
	if !ok {
		return
	}
	holder.Entry = logger.WithFields(fields)
}

func toContext(ctx context.Context, entry *logrus.Entry) context.Context {
	h, ok := ctx.Value(grpc_logging.InternalContextMarker).(*holder)
	if !ok {
		return context.WithValue(ctx, grpc_logging.InternalContextMarker, &holder{entry})
	}
	h.Entry = entry
	return ctx
}
