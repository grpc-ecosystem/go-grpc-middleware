// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package tracing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

func TestTagsCarrier_Set_JaegerTraceFormat(t *testing.T) {
	var (
		fakeTraceSampled   = 1
		fakeInboundTraceId = "deadbeef"
		fakeInboundSpanId  = "c0decafe"
		traceHeaderName    = "uber-trace-id"
	)

	traceHeaderValue := fmt.Sprintf("%s:%s:%s:%d", fakeInboundTraceId, fakeInboundSpanId, fakeInboundSpanId, fakeTraceSampled)

	c := &tagsCarrier{
		Tags:            tags.NewTags(),
		traceHeaderName: traceHeaderName,
	}

	c.Set(traceHeaderName, traceHeaderValue)

	assert.EqualValues(t, map[string]string{
		TagTraceId: fakeInboundTraceId,
		TagSpanId:  fakeInboundSpanId,
		TagSampled: "true",
	}, c.Tags.Values())
}
