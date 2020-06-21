package external

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

// Test to ensure TagBasedRequestFieldExtractor does not panic when encountering private struct members such as
// when using gogoproto.stdtime which results in a time.Time that has private struct members
func TestTaggedRequestFiledExtractor_GogoTime(t *testing.T) {
	ts := time.Date(2010, 01, 01, 0, 0, 0, 0, time.UTC)
	req := &gogofieldstestpb.GoGoProtoStdTime{
		Timestamp: &ts,
	}
	assert.NotPanics(t, func() {
		valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
		assert.Empty(t, valMap)
	})
}
func TestTaggedRequestFiledExtractor_GogoPingRequest(t *testing.T) {
	req := &gogofieldstestpb.PingRequest{
		Ping: &gogofieldstestpb.Ping{
			Id: &gogofieldstestpb.PingId{
				Id: 1337, // logfield is ping_id
			},
			Value: "something",
		},
		Meta: &gogofieldstestpb.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
	assert.EqualValues(t, "1337", valMap["ping_id"])
	assert.EqualValues(t, "[tagone tagtwo]", valMap["meta_tags"])
}

func TestTaggedRequestFiledExtractor_GogoPongRequest(t *testing.T) {
	req := &gogofieldstestpb.PongRequest{
		Pong: &gogofieldstestpb.Pong{
			Id: "some_id",
		},
		Meta: &gogofieldstestpb.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
	assert.EqualValues(t, "some_id", valMap["pong_id"])
	assert.EqualValues(t, "[tagone tagtwo]", valMap["meta_tags"])
}
