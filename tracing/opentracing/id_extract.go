package grpc_opentracing

import (
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/grpclog"
)

const (
	TagTraceId           = "trace.traceid"
	TagSpanId            = "trace.spanid"
	JaegerNotSampledFlag = "0"
)

// injectOpentracingIdsToTags writes the given context to the ctxtags if the trace is sampled.
// This is done in an incredibly hacky way, because the public-facing interface of opentracing doesn't give access to
// the TraceId and SpanId of the SpanContext. Only the Tracer's Inject/Extract methods know what these are.
// Most tracers have them encoded as keys with 'traceid' and 'spanid':
// https://github.com/openzipkin/zipkin-go-opentracing/blob/594640b9ef7e5c994e8d9499359d693c032d738c/propagation_ot.go#L29
// https://github.com/opentracing/basictracer-go/blob/1b32af207119a14b1b231d451df3ed04a72efebf/propagation_ot.go#L26
// Jaeger from Uber use one-key schema with next format '{trace-id}:{span-id}:{parent-span-id}:{flags}'
// https://www.jaegertracing.io/docs/client-libraries/#trace-span-identity
func injectOpentracingIdsToTags(span opentracing.Span, tags grpc_ctxtags.Tags) {
	carrier := &tagsCarrier{}
	if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
		grpclog.Infof("grpc_opentracing: failed extracting trace info into ctx %v", err)
	}

	if carrier.sampled {
		tags.Set(TagTraceId, carrier.traceID)
		tags.Set(TagSpanId, carrier.spanID)
	}
}

// tagsCarrier is a really hacky way of
type tagsCarrier struct {
	sampled bool
	traceID string
	spanID  string
}

func (t *tagsCarrier) Set(key, val string) {
	if strings.Contains(strings.ToLower(key), "traceid") {
		t.traceID = val
	}

	if strings.Contains(strings.ToLower(key), "spanid") && !strings.Contains(strings.ToLower(key), "parent") {
		t.spanID = val
	}

	if strings.Contains(strings.ToLower(key), "sampled") && val == "true" {
		t.sampled = true
	}

	if key == "uber-trace-id" {
		parts := strings.Split(val, ":")
		if len(parts) == 4 {
			if parts[3] != JaegerNotSampledFlag {
				t.traceID = parts[0]
				t.spanID = parts[1]
				t.sampled = true
			}
		}
	}
}
