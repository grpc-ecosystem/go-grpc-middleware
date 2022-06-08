// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
)

type span struct {
	span trace.Span
	ctx  context.Context
}

var _ tracing.Span = (*span)(nil)

func newSpan(ctx context.Context, rawSpan trace.Span) *span {
	return &span{
		ctx:  ctx,
		span: rawSpan,
	}
}

func (s *span) End() {
	s.span.End()
}

func (s *span) SetStatus(code grpccodes.Code, message string) {
	if code != grpccodes.OK {
		s.span.SetStatus(codes.Error, message)
	} else {
		s.span.SetStatus(codes.Ok, message)
	}
}

func (s *span) SetAttributes(keyvals ...interface{}) {
	s.span.SetAttributes(translateKeyValue(keyvals)...)
}

func (s *span) AddEvent(name string, keyvals ...interface{}) {
	attributes := translateKeyValue(keyvals)

	s.span.AddEvent(name, trace.WithAttributes(attributes...))
}

func translateKeyValue(keyvals ...interface{}) []attribute.KeyValue {
	if len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}

	attributes := make([]attribute.KeyValue, 0, len(keyvals)/2+1)

	for i := 0; i < len(keyvals); i += 2 {
		k, keyOK := keyvals[i].(string)
		if !keyOK {
			// Skip this key
			// TODO: should we log something to warn users?
			continue
		}

		switch v := keyvals[i+1].(type) {
		case bool:
			attributes = append(attributes, attribute.Bool(k, v))
		case int:
			attributes = append(attributes, attribute.Int(k, v))
		case int32:
			attributes = append(attributes, attribute.Int(k, int(v)))
		case int64:
			attributes = append(attributes, attribute.Int64(k, v))
		case float32:
			attributes = append(attributes, attribute.Float64(k, float64(v)))
		case float64:
			attributes = append(attributes, attribute.Float64(k, v))
		case string:
			attributes = append(attributes, attribute.String(k, v))
		default:
			// Unsupported type
			// TODO: should we log something to warn users?
			continue
		}
	}

	return attributes
}
