// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWrapper(t *testing.T) {
	ctx := context.Background()

	require.NoError(t, validate(ctx, testpb.GoodPing, false, nil))
	require.Error(t, validate(ctx, testpb.BadPing, false, nil))
	require.NoError(t, validate(ctx, testpb.GoodPing, true, nil))
	require.Error(t, validate(ctx, testpb.BadPing, true, nil))

	require.NoError(t, validate(ctx, testpb.GoodPingError, false, nil))
	require.Error(t, validate(ctx, testpb.BadPingError, false, nil))
	require.NoError(t, validate(ctx, testpb.GoodPingError, true, nil))
	require.Error(t, validate(ctx, testpb.BadPingError, true, nil))

	assert.NoError(t, validate(ctx, testpb.GoodPingResponse, false, nil))
	assert.NoError(t, validate(ctx, testpb.GoodPingResponse, true, nil))
	require.Error(t, validate(ctx, testpb.BadPingResponse, false, nil))
	assert.Error(t, validate(ctx, testpb.BadPingResponse, true, nil))
}
