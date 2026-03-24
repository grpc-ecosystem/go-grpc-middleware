// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

// TestLevelValues verifies that Level constants retain their expected numeric
// values after the type was changed from a bare int to an alias of slog.Level.
// Any change here would be a breaking change for users who rely on the numeric
// values (e.g. serialization, comparison, or custom code-to-level mappers).
func TestLevelValues(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		wantInt int
		slogLvl slog.Level
	}{
		{name: "LevelDebug", level: LevelDebug, wantInt: -4, slogLvl: slog.LevelDebug},
		{name: "LevelInfo", level: LevelInfo, wantInt: 0, slogLvl: slog.LevelInfo},
		{name: "LevelWarn", level: LevelWarn, wantInt: 4, slogLvl: slog.LevelWarn},
		{name: "LevelError", level: LevelError, wantInt: 8, slogLvl: slog.LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Numeric value must remain unchanged (backward compatibility).
			assert.Equal(t, tt.wantInt, int(tt.level), "numeric value of %s must be %d", tt.name, tt.wantInt)
			// Must equal the corresponding slog.Level constant.
			assert.Equal(t, tt.slogLvl, slog.Level(tt.level), "%s must equal slog.%s", tt.name, tt.slogLvl)
		})
	}
}

// TestLevelSlogRoundTrip ensures Level to slog.Level conversions are lossless.
func TestLevelSlogRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{name: "LevelDebug", level: LevelDebug},
		{name: "LevelInfo", level: LevelInfo},
		{name: "LevelWarn", level: LevelWarn},
		{name: "LevelError", level: LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slogLevel := slog.Level(tt.level)
			backToLevel := Level(slogLevel)
			assert.Equal(t, tt.level, backToLevel, "round-trip conversion must be lossless for %v", tt.level)
		})
	}
}

// TestLevelComparison verifies the ordering contract: Debug < Info < Warn < Error.
func TestLevelComparison(t *testing.T) {
	tests := []struct {
		name   string
		lower  Level
		higher Level
	}{
		{name: "Debug < Info", lower: LevelDebug, higher: LevelInfo},
		{name: "Info < Warn", lower: LevelInfo, higher: LevelWarn},
		{name: "Warn < Error", lower: LevelWarn, higher: LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Less(t, tt.lower, tt.higher)
		})
	}
}

// TestLoggerFuncReceivesCorrectLevel asserts that when the logging interceptor
// invokes Logger.Log, the level argument is still the expected Level type and
// value. This catches accidental type mismatches that could arise after the
// Level alias change.
func TestLoggerFuncReceivesCorrectLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{name: "LevelDebug", level: LevelDebug},
		{name: "LevelInfo", level: LevelInfo},
		{name: "LevelWarn", level: LevelWarn},
		{name: "LevelError", level: LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured Level
			fn := LoggerFunc(func(_ context.Context, lvl Level, _ string, _ ...any) {
				captured = lvl
			})
			fn.Log(context.Background(), tt.level, "test message")
			assert.Equal(t, tt.level, captured)
		})
	}
}

// TestLoggerInterfaceAcceptsSlogLevel ensures that a slog.Level value can be
// passed through the Logger interface after a simple cast, which is the primary
// use-case unlocked by the alias change.
func TestLoggerInterfaceAcceptsSlogLevel(t *testing.T) {
	tests := []struct {
		name      string
		slogLevel slog.Level
		wantLevel Level
	}{
		{name: "slog.LevelDebug", slogLevel: slog.LevelDebug, wantLevel: LevelDebug},
		{name: "slog.LevelInfo", slogLevel: slog.LevelInfo, wantLevel: LevelInfo},
		{name: "slog.LevelWarn", slogLevel: slog.LevelWarn, wantLevel: LevelWarn},
		{name: "slog.LevelError", slogLevel: slog.LevelError, wantLevel: LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured Level
			fn := LoggerFunc(func(_ context.Context, lvl Level, _ string, _ ...any) {
				captured = lvl
			})
			fn.Log(context.Background(), Level(tt.slogLevel), "msg")
			assert.Equal(t, tt.wantLevel, captured,
				"Level(%v) passed to Logger should equal %v", tt.slogLevel, tt.wantLevel)
		})
	}
}

// TestDefaultServerCodeToLevel_LevelTypes confirms that the server code-to-level
// mapper returns values that are valid slog.Level equivalents.
func TestDefaultServerCodeToLevel_LevelTypes(t *testing.T) {
	tests := []struct {
		code    codes.Code
		wantLvl Level
	}{
		{codes.OK, LevelInfo},
		{codes.NotFound, LevelInfo},
		{codes.InvalidArgument, LevelInfo},
		{codes.Internal, LevelError},
		{codes.Unknown, LevelError},
		{codes.DataLoss, LevelError},
		{codes.Unauthenticated, LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			got := DefaultServerCodeToLevel(tt.code)
			assert.Equal(t, tt.wantLvl, got)
			// Verify the returned Level is a valid slog.Level.
			assert.Equal(t, slog.Level(tt.wantLvl), slog.Level(got))
		})
	}
}

// TestDefaultClientCodeToLevel_LevelTypes confirms that the client code-to-level
// mapper returns values that are valid slog.Level equivalents.
func TestDefaultClientCodeToLevel_LevelTypes(t *testing.T) {
	tests := []struct {
		code    codes.Code
		wantLvl Level
	}{
		{codes.OK, LevelDebug},
		{codes.Canceled, LevelDebug},
		{codes.NotFound, LevelDebug},
		{codes.Unknown, LevelInfo},
		{codes.DeadlineExceeded, LevelInfo},
		{codes.Unauthenticated, LevelInfo},
		{codes.Internal, LevelWarn},
		{codes.Unavailable, LevelWarn},
	}
	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			got := DefaultClientCodeToLevel(tt.code)
			assert.Equal(t, tt.wantLvl, got)
			// Verify the returned Level is a valid slog.Level.
			assert.Equal(t, slog.Level(tt.wantLvl), slog.Level(got))
		})
	}
}

// TestCustomCodeToLevelWithSlogLevel simulates a user-defined CodeToLevel
// function that returns slog.Level values cast to Level. This is the main
// integration pattern enabled by the alias change.
func TestCustomCodeToLevelWithSlogLevel(t *testing.T) {
	// A user-defined mapper using slog.Level values directly.
	customMapper := func(code codes.Code) Level {
		switch code {
		case codes.OK:
			return Level(slog.LevelDebug)
		case codes.Internal:
			return Level(slog.LevelError)
		default:
			return Level(slog.LevelInfo)
		}
	}

	require.Equal(t, LevelDebug, customMapper(codes.OK))
	require.Equal(t, LevelError, customMapper(codes.Internal))
	require.Equal(t, LevelInfo, customMapper(codes.NotFound))
}

// TestLevelZeroValue ensures the zero value of Level equals LevelInfo (which is
// slog.LevelInfo = 0). This is important because Go zero-initializes variables
// and this was always the implicit contract.
func TestLevelZeroValue(t *testing.T) {
	var zeroLevel Level
	assert.Equal(t, LevelInfo, zeroLevel, "zero value of Level must be LevelInfo")
	assert.Equal(t, slog.LevelInfo, slog.Level(zeroLevel), "zero value must map to slog.LevelInfo")
}
