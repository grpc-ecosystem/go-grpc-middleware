// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package tags_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

// TODO(bwplotka): Add tests/examples https://github.com/grpc-ecosystem/go-grpc-middleware/issues/382
func TestCodeGenRequestLogFieldExtractor_ManualIsDeclared(t *testing.T) {
	req := &testpb.PingRequest{Value: "my_value"}
	valMap := tags.CodeGenRequestFieldExtractor("", req)
	require.Len(t, valMap, 1, "PingRequest should have a ExtractLogFields method declared in test.manual_extractfields.pb")
	require.EqualValues(t, valMap, map[string]string{"value": "my_value"})
}
