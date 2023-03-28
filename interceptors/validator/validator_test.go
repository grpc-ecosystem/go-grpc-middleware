// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
)

type TestLogger struct{}

func (l *TestLogger) Log(lvl logging.Level, msg string) {}

func (l *TestLogger) With(fields ...string) logging.Logger {
	return &TestLogger{}
}

func TestValidateWrapper(t *testing.T) {
	assert.NoError(t, validate(testpb.GoodPing, false, logging.ERROR, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPing, false, logging.ERROR, &TestLogger{}))
	assert.NoError(t, validate(testpb.GoodPing, true, logging.ERROR, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPing, true, logging.ERROR, &TestLogger{}))

	assert.NoError(t, validate(testpb.GoodPingError, false, logging.ERROR, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingError, false, logging.ERROR, &TestLogger{}))
	assert.NoError(t, validate(testpb.GoodPingError, true, logging.ERROR, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingError, true, logging.ERROR, &TestLogger{}))

	assert.NoError(t, validate(testpb.GoodPingResponse, false, logging.ERROR, &TestLogger{}))
	assert.NoError(t, validate(testpb.GoodPingResponse, true, logging.ERROR, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingResponse, false, logging.ERROR, &TestLogger{}))
	assert.Error(t, validate(testpb.BadPingResponse, true, logging.ERROR, &TestLogger{}))
}
