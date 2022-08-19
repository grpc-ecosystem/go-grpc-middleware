// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logr_test

import (
	"testing"

	grpclogr "github.com/grpc-ecosystem/go-grpc-middleware/providers/logr/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/stretchr/testify/assert"
	"k8s.io/klog/v2/ktesting"
)

type (
	Entry struct {
		Message    string
		Verbosity  int
		WithKVList []interface{}
	}
)

func mapKtesingEntries(underlierLogger ktesting.Underlier) (entries []Entry) {
	for _, entry := range underlierLogger.GetBuffer().Data() {
		entries = append(entries, Entry{
			Message:    entry.Message,
			Verbosity:  entry.Verbosity,
			WithKVList: entry.WithKVList,
		})
	}

	return entries
}

func TestLogger_Log(t *testing.T) {
	testLogger := ktesting.NewLogger(t, ktesting.NewConfig())
	underlierLogger := testLogger.GetSink().(ktesting.Underlier)
	logger := grpclogr.InterceptorLogger(testLogger)

	loggerWithFields := logger.With("key-1", "value-1")
	loggerWithFields.Log(logging.DEBUG, "debug message")
	logger.Log(logging.INFO, "some info message")
	loggerWithFields.With("key-2", "value-2", "key-3", "value-3").Log(logging.WARNING, "warn")
	logger.Log(logging.ERROR, "error")

	assert.Equal(t, []Entry{
		{
			Message:    "debug message",
			Verbosity:  4,
			WithKVList: []interface{}{"key-1", "value-1"},
		},
		{
			Message:   "some info message",
			Verbosity: 2,
		},
		{
			Message:    "warn",
			Verbosity:  1,
			WithKVList: []interface{}{"key-1", "value-1", "key-2", "value-2", "key-3", "value-3"},
		},
		{
			Message:   "error",
			Verbosity: 0,
		},
	}, mapKtesingEntries(underlierLogger))
}
