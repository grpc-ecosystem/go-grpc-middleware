// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
)

func TestValidateWrapper(t *testing.T) {
	ctx := context.Background()

	assert.NoError(t, validate(ctx, testpb.GoodPing, false, nil))
	assert.Error(t, validate(ctx, testpb.BadPing, false, nil))
	assert.NoError(t, validate(ctx, testpb.GoodPing, true, nil))
	assert.Error(t, validate(ctx, testpb.BadPing, true, nil))

	assert.NoError(t, validate(ctx, testpb.GoodPingError, false, nil))
	assert.Error(t, validate(ctx, testpb.BadPingError, false, nil))
	assert.NoError(t, validate(ctx, testpb.GoodPingError, true, nil))
	assert.Error(t, validate(ctx, testpb.BadPingError, true, nil))

	assert.NoError(t, validate(ctx, testpb.GoodPingResponse, false, nil))
	assert.NoError(t, validate(ctx, testpb.GoodPingResponse, true, nil))
	assert.Error(t, validate(ctx, testpb.BadPingResponse, false, nil))
	assert.Error(t, validate(ctx, testpb.BadPingResponse, true, nil))
}
