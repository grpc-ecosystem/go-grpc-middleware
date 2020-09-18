// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package tags_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/gogotestpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

func TestCodeGenRequestLogFieldExtractor_ManualIsDeclared(t *testing.T) {
	req := &testpb.PingRequest{Value: "my_value"}
	valMap := tags.CodeGenRequestFieldExtractor("", req)
	require.Len(t, valMap, 1, "PingRequest should have a ExtractLogFields method declared in test.manual_extractfields.pb")
	require.EqualValues(t, valMap, map[string]string{"value": "my_value"})
}

func TestTaggedRequestFiledExtractor_PingRequest(t *testing.T) {
	req := &gogotestpb.PingRequest{
		Ping: &gogotestpb.Ping{
			Id: &gogotestpb.PingId{
				Id: 1337, // logfield is ping_id
			},
			Value: "something",
		},
		Meta: &gogotestpb.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
	assert.EqualValues(t, "1337", valMap["ping_id"])
	assert.EqualValues(t, "[tagone tagtwo]", valMap["meta_tags"])
}

func TestTaggedRequestFiledExtractor_PongRequest(t *testing.T) {
	req := &gogotestpb.PongRequest{
		Pong: &gogotestpb.Pong{
			Id: "some_id",
		},
		Meta: &gogotestpb.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
	assert.EqualValues(t, "some_id", valMap["pong_id"])
	assert.EqualValues(t, "[tagone tagtwo]", valMap["meta_tags"])
}

func TestTaggedRequestFiledExtractor_OneOfLogField(t *testing.T) {
	req := &gogotestpb.OneOfLogField{
		Identifier: &gogotestpb.OneOfLogField_BarId{
			BarId: "bar-log-field",
		},
	}
	valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
	assert.EqualValues(t, "bar-log-field", valMap["bar_id"])
}

// Test to ensure TagBasedRequestFieldExtractor does not panic when encountering private struct members such as
// when using gogoproto.stdtime which results in a time.Time that has private struct members
func TestTaggedRequestFiledExtractor_GogoTime(t *testing.T) {
	ts := time.Date(2010, 01, 01, 0, 0, 0, 0, time.UTC)
	req := &gogotestpb.GoGoProtoStdTime{
		Timestamp: &ts,
	}
	assert.NotPanics(t, func() {
		valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
		assert.Empty(t, valMap)
	})
}
