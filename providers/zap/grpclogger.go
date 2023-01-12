// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package zap

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "grpc")
)

// ReplaceGrpcLoggerV2 replaces the grpclog.LoggerV2 with the provided logger.
// It should be called before any gRPC functions. Logging verbosity defaults to info level.
// To adjust gRPC logging verbosity, see ReplaceGrpcLoggerV2WithVerbosity.
func ReplaceGrpcLoggerV2(logger *zap.Logger) {
	ReplaceGrpcLoggerV2WithVerbosity(logger, 0)
}

// ReplaceGrpcLoggerV2WithVerbosity replaces the grpclog.Logger with the provided logger and verbosity.
// It should be called before any gRPC functions.
// verbosity correlates to grpclogs verbosity levels. A higher verbosity value results in less logging.
func ReplaceGrpcLoggerV2WithVerbosity(logger *zap.Logger, verbosity int) {
	zgl := &zapGrpcLoggerV2{
		logger:    logger.With(SystemField, zap.Bool("grpc_log", true)).WithOptions(zap.AddCallerSkip(2)),
		verbosity: verbosity,
	}
	grpclog.SetLoggerV2(zgl)
}

type zapGrpcLoggerV2 struct {
	logger    *zap.Logger
	verbosity int
}

func (l *zapGrpcLoggerV2) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Infoln(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) InfoDepth(depth int, args ...interface{}) {
	l.logger.WithOptions(zap.AddCallerSkip(depth)).Info(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Warning(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Warningln(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) WarningDepth(depth int, args ...interface{}) {
	l.logger.WithOptions(zap.AddCallerSkip(depth)).Warn(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Errorln(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) ErrorDepth(depth int, args ...interface{}) {
	l.logger.WithOptions(zap.AddCallerSkip(depth)).Error(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) FatalDepth(depth int, args ...interface{}) {
	l.logger.WithOptions(zap.AddCallerSkip(depth)).Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) V(level int) bool {
	// Check whether the verbosity of the current log ('level') is within the specified threshold ('l.verbosity').
	// As in https://github.com/grpc/grpc-go/blob/41e044e1c82fcf6a5801d6cbd7ecf952505eecb1/grpclog/loggerv2.go#L199-L201.
	return level <= l.verbosity
}
