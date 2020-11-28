package grpc_zap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/grpclog"
)

func Test_zapGrpcLogger_V(t *testing.T) {
	// copied from gRPC
	const (
		// infoLog indicates Info severity.
		infoLog int = iota
		// warningLog indicates Warning severity.
		warningLog
		// errorLog indicates Error severity.
		errorLog
		// fatalLog indicates Fatal severity.
		fatalLog
	)

	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	ReplaceGrpcLoggerV2WithVerbosity(logger, warningLog)
	assert.False(t, grpclog.V(infoLog))
	assert.True(t, grpclog.V(warningLog))
	assert.True(t, grpclog.V(errorLog))
	assert.True(t, grpclog.V(fatalLog))
}
