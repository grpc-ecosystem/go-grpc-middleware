// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package ctx_glog

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"golang.org/x/net/context"
)

type ctxLoggerMarker struct{}

type ctxLogger struct {
	logger *Entry
	fields Fields
}

var (
	ctxLoggerKey = &ctxLoggerMarker{}
)

// AddFields adds glog fields to the logger.
func AddFields(ctx context.Context, fields Fields) {
	l, ok := ctx.Value(ctxLoggerKey).(*ctxLogger)
	if !ok || l == nil {
		return
	}
	for k, v := range fields {
		l.fields[k] = v
	}
}

// Extract takes the call-scoped Entry from ctx_glog middleware.
//
// If the ctx_glog middleware wasn't used, a no-op `Entry` is returned. This makes it safe to
// use regardless.
func Extract(ctx context.Context) *Entry {
	l, ok := ctx.Value(ctxLoggerKey).(*ctxLogger)
	if !ok || l == nil {
		return NewEntry(nullLogger)
	}

	fields := Fields{}

	// Add grpc_ctxtags tags metadata until now.
	tags := grpc_ctxtags.Extract(ctx)
	for k, v := range tags.Values() {
		fields[k] = v
	}

	// Add glog fields added until now.
	for k, v := range l.fields {
		fields[k] = v
	}

	return l.logger.WithFields(fields)
}

// ToContext adds the Entry to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, entry *Entry) context.Context {
	l := &ctxLogger{
		logger: entry,
		fields: Fields{},
	}
	return context.WithValue(ctx, ctxLoggerKey, l)
}
