// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentelemetry

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

func newServerSpanFromContext(ctx context.Context, fullMethod string, tracer trace.Tracer, propagators propagation.TextMapPropagator) (context.Context, trace.Span) {
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()

	b, spanCtx := otelgrpc.Extract(ctx, &metadataCopy, otelgrpc.WithPropagators(propagators))
	ctx = baggage.ContextWithBaggage(ctx, b)

	name, attrs := spanInfo(ctx, fullMethod)
	ctx, span := tracer.Start(
		trace.ContextWithRemoteSpanContext(ctx, spanCtx),
		name,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...),
	)

	return ctx, span
}
