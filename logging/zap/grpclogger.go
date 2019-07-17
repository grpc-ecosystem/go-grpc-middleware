// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zap

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets the given zap.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
// Deprecated: use ReplaceGrpcLoggerV2.
func ReplaceGrpcLogger(logger *zap.Logger) {
	zgl := &zapGrpcLogger{logger.With(SystemField, zap.Bool("grpc_log", true))}
	grpclog.SetLogger(zgl)
}

type zapGrpcLogger struct {
	logger *zap.Logger
}

func (l *zapGrpcLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Print(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Printf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) Println(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

// ReplaceGrpcLoggerV2 replaces the grpc_log.LoggerV2 with the provided logger.
func ReplaceGrpcLoggerV2(logger *zap.Logger) {
	ReplaceGrpcLoggerV2WithVerbosity(logger, 0)
}

// ReplaceGrpcLoggerV2WithVerbosity replaces the grpc_.LoggerV2 with the provided logger and verbosity.
func ReplaceGrpcLoggerV2WithVerbosity(logger *zap.Logger, verbosity int) {
	zgl := &zapGrpcLoggerV2{
		logger:    logger.With(SystemField, zap.Bool("grpc_log", true)),
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

func (l *zapGrpcLoggerV2) Warning(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Warningln(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
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

func (l *zapGrpcLoggerV2) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) V(level int) bool {
	return level <= l.verbosity
}
