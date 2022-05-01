package opentracing

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
)

type span struct {
	span opentracing.Span
	ctx  context.Context

	initial bool
}

// Compatibility check.
var _ tracing.Span = &span{}

func newSpan(rawSpan opentracing.Span, ctx context.Context) *span {
	return &span{
		span:    rawSpan,
		ctx:     ctx,
		initial: true,
	}
}

func (s *span) End() {
	// Middleware tags only record once
	if s.initial {
		s.initial = false

		iter := logging.ExtractFields(s.ctx).Iter()
		for iter.Next() {
			k, v := iter.At()
			s.span.SetTag(k, v)
		}
	}

	s.span.Finish()
}

func (s *span) SetStatus(code codes.Code, msg string) {
	if code != codes.OK {
		ext.Error.Set(s.span, true)
		s.span.LogFields(log.String("event", "error"), log.String("message", msg))
	}
}

func (s *span) AddEvent(name string, keyvals ...interface{}) {
	if len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}

	fields := make([]log.Field, 0, len(keyvals)/2+1)

	fields = append(fields, log.String("event", name))

	for i := 0; i < len(keyvals); i += 2 {
		k, keyOK := keyvals[i].(string)
		if !keyOK {
			// Skip this key
			// TODO: should we log something to warn users?
			continue
		}

		switch v := keyvals[i+1].(type) {
		case bool:
			fields = append(fields, log.Bool(k, v))
		case int:
			fields = append(fields, log.Int(k, v))
		case int32:
			fields = append(fields, log.Int32(k, v))
		case int64:
			fields = append(fields, log.Int64(k, v))
		case uint32:
			fields = append(fields, log.Uint32(k, v))
		case uint64:
			fields = append(fields, log.Uint64(k, v))
		case float32:
			fields = append(fields, log.Float32(k, v))
		case float64:
			fields = append(fields, log.Float64(k, v))
		case string:
			fields = append(fields, log.String(k, v))
		default:
			// Unsupported type
			// TODO: should we log something to warn users?
			continue
		}
	}

	s.span.LogFields(fields...)
}

func (s *span) SetAttributes(keyvals ...interface{}) {
	if len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}

	for i := 0; i < len(keyvals); i += 2 {
		k, keyOK := keyvals[i].(string)
		if !keyOK {
			// Skip this key
			// TODO: should we log something to warn users?
			continue
		}

		v := keyvals[i+1]
		if v != nil {
			s.span.SetTag(k, v)
		}
	}
}
