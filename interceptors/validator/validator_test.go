// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
)

type TestLogger struct{}

func (l *TestLogger) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {}

func TestValidateWrapper(t *testing.T) {
	assert.NoError(t, validate(testpb.GoodPing, false, logging.LevelError, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPing, false, logging.LevelError, &TestLogger{}))
	assert.NoError(t, validate(testpb.GoodPing, true, logging.LevelError, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPing, true, logging.LevelError, &TestLogger{}))

	assert.NoError(t, validate(testpb.GoodPingError, false, logging.LevelError, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingError, false, logging.LevelError, &TestLogger{}))
	assert.NoError(t, validate(testpb.GoodPingError, true, logging.LevelError, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingError, true, logging.LevelError, &TestLogger{}))

	assert.NoError(t, validate(testpb.GoodPingResponse, false, logging.LevelError, &TestLogger{}))
	assert.NoError(t, validate(testpb.GoodPingResponse, true, logging.LevelError, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingResponse, false, logging.LevelError, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingResponse, true, logging.LevelError, &TestLogger{}))
}
