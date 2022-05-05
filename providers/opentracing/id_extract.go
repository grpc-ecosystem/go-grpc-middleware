// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package opentracing

import (
	"strings"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/grpclog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

const (
	TagTraceId           = "trace.traceid"
	TagSpanId            = "trace.spanid"
	TagSampled           = "trace.sampled"
	jaegerNotSampledFlag = "0"
)

// injectOpentracingIdsToTags writes trace data to tags.
// This is done in an incredibly hacky way, because the public-facing interface of opentracing doesn't give access to
// the TraceId and SpanId of the SpanContext. Only the Tracer's Inject/Extract methods know what these are.
// Most tracers have them encoded as keys with 'traceid' and 'spanid':
// https://github.com/openzipkin/zipkin-go-opentracing/blob/594640b9ef7e5c994e8d9499359d693c032d738c/propagation_ot.go#L29
// https://github.com/opentracing/basictracer-go/blob/1b32af207119a14b1b231d451df3ed04a72efebf/propagation_ot.go#L26
// Jaeger from Uber use one-key schema with next format '{trace-id}:{span-id}:{parent-span-id}:{flags}'
// https://www.jaegertracing.io/docs/client-libraries/#trace-span-identity
// Datadog uses keys ending with 'trace-id' and 'parent-id' (for span) by default:
// https://github.com/DataDog/dd-trace-go/blob/v1/ddtrace/tracer/textmap.go#L77
func injectOpentracingIdsToTags(traceHeaderName string, span opentracing.Span, fields logging.Fields) logging.Fields {
	tagsCarrier := tagsCarrier{fields: fields, traceHeaderName: traceHeaderName}
	if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders,
		&tagsCarrier); err != nil {
		grpclog.Infof("grpc_opentracing: failed extracting trace info into ctx %v", err)
	}
	return tagsCarrier.fields
}

// tagsCarrier is a really hacky way of
type tagsCarrier struct {
	fields          logging.Fields
	traceHeaderName string
}

func (t *tagsCarrier) Set(key, val string) {
	key = strings.ToLower(key)
	if strings.Contains(key, "traceid") {
		t.fields = append(t.fields, TagTraceId, val) // this will most likely be base-16 (hex) encoded
	}

	if strings.Contains(key, "spanid") && !strings.Contains(strings.ToLower(key), "parent") {
		t.fields = append(t.fields, TagSpanId, val) // this will most likely be base-16 (hex) encoded
	}

	if strings.Contains(key, "sampled") {
		switch val {
		case "true", "false":
			t.fields = append(t.fields, TagSampled, val)
		}
	}

	if key == t.traceHeaderName {
		parts := strings.Split(val, ":")
		if len(parts) == 4 {
			t.fields = append(t.fields, TagTraceId, parts[0], TagSpanId, parts[1])

			if parts[3] != jaegerNotSampledFlag {
				t.fields = append(t.fields, TagSampled, "true")
			} else {
				t.fields = append(t.fields, TagSampled, "false")
			}
		}
	}

	if strings.HasSuffix(key, "trace-id") {
		t.fields = append(t.fields, TagTraceId, val)
	}

	if strings.HasSuffix(key, "parent-id") {
		t.fields = append(t.fields, TagSpanId, val)
	}
}
