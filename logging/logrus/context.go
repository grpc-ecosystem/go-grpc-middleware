// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logrus

import (
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var (
	ctxMarker = "logrus-logger"
)

type holder struct {
	*logrus.Entry
}

// Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.
//
// If the grpc_logrus middleware wasn't used, a no-op `logrus.Entry` is returned. This makes it safe to
// use regardless.
func Extract(ctx context.Context) *logrus.Entry {
	h, ok := ctx.Value(ctxMarker).(*holder)
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
	holder, ok := ctx.Value(ctxMarker).(*holder)
	if !ok {
		return
	}
	holder.Entry = logger.WithFields(fields)
}

func toContext(ctx context.Context, entry *logrus.Entry) context.Context {
	h, ok := ctx.Value(ctxMarker).(*holder)
	if !ok {
		return context.WithValue(ctx, ctxMarker, &holder{entry})
	}
	h.Entry = entry
	return ctx
}
