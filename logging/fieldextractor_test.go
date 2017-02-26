// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logging_test

import "testing"
import (
	"github.com/mwitkow/go-grpc-middleware/logging"
	pb_gogotestproto "github.com/mwitkow/go-grpc-middleware/testing/gogotestproto"
	pb_testproto "github.com/mwitkow/go-grpc-middleware/testing/testproto"

	"github.com/stretchr/testify/require"
)

func TestCodeGenRequestLogFieldExtractor_ManualIsDeclared(t *testing.T) {
	req := &pb_testproto.PingRequest{Value: "my_value"}
	keys, values := grpc_logging.CodeGenRequestLogFieldExtractor("", req)
	require.Len(t, keys, 1, "PingRequest should have a ExtractLogFields method declared in test.manual_extractfields.pb")
	require.EqualValues(t, []string{"request.value"}, keys)
	require.EqualValues(t, []interface{}{"my_value"}, values)
}

func TestTagedRequestFiledExtractor_PingRequest(t *testing.T) {
	req := &pb_gogotestproto.PingRequest{
		Ping: &pb_gogotestproto.Ping{
			Id: &pb_gogotestproto.PingId{
				Id: 1337, // logfield is ping_id
			},
			Value: "something",
		},
		Meta: &pb_gogotestproto.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	keys, values := grpc_logging.TagedRequestFiledExtractor("", req)
	require.EqualValues(t, []string{"ping_id", "meta_tags"}, keys)
	require.EqualValues(t, []interface{}{int32(1337), []string{"tagone", "tagtwo"}}, values)
}

func TestTagedRequestFiledExtractor_PongRequest(t *testing.T) {
	req := &pb_gogotestproto.PongRequest{
		Pong: &pb_gogotestproto.Pong{
			Id: "some_id",
		},
		Meta: &pb_gogotestproto.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	keys, values := grpc_logging.TagedRequestFiledExtractor("", req)
	require.EqualValues(t, []string{"pong_id", "meta_tags"}, keys)
	require.EqualValues(t, []interface{}{"some_id", []string{"tagone", "tagtwo"}}, values)
}
