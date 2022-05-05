// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/grpclog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
)

func newServerSpanFromInbound(ctx context.Context, tracer opentracing.Tracer, traceHeaderName, fullMethodName string) (context.Context, opentracing.Span) {
	md := metadata.ExtractIncoming(ctx)
	parentSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, metadataTextMap(md))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		grpclog.Infof("grpc_opentracing: failed parsing trace information: %v", err)
	}

	serverSpan := tracer.StartSpan(
		fullMethodName,
		// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
		ext.RPCServerOption(parentSpanContext),
		grpcTag,
	)

	// Log context information.
	iter := logging.ExtractFields(ctx).Iter()
	for iter.Next() {
		k, v := iter.At()
		serverSpan.SetTag(k, v)
	}

	fields := injectOpentracingIdsToTags(traceHeaderName, serverSpan, logging.ExtractFields(ctx))
	ctx = logging.InjectFields(ctx, fields)
	return opentracing.ContextWithSpan(ctx, serverSpan), serverSpan
}
