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
// Deprecated: use ReplaceGrpcLoggerV2
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

type GrpcLoggerV2Options struct {
	verbosity int
}

// DefaultGrpcLoggerV2Options is the default options for ReplaceGrpcLoggerV2
var DefaultGrpcLoggerV2Options = &GrpcLoggerV2Options{}

// ReplaceGrpcLoggerV2 replaces the grpc_log.LoggerV2 with the input zap.Logger
// This logger allows you to set grpc verbosity level
func ReplaceGrpcLoggerV2(logger *zap.Logger, opts *GrpcLoggerV2Options) {
	zgl := &zapGrpcLoggerV2{Logger: logger.With(SystemField, zap.Bool("grpc_log", true))}

	if opts != nil {
		zgl.verbosity = opts.verbosity
	}

	grpclog.SetLoggerV2(zgl)
}

const (
	errorLevel = iota
	warnLevel
	infoLevel
)

type zapGrpcLoggerV2 struct {
	*zap.Logger
	verbosity int
	severity  int
}

func (l *zapGrpcLoggerV2) Info(args ...interface{}) {
	if l.severity >= infoLevel {
		l.Logger.Info(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Infoln(args ...interface{}) {
	if l.severity >= infoLevel {
		l.Logger.Info(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Infof(format string, args ...interface{}) {
	if l.severity >= infoLevel {
		l.Logger.Info(fmt.Sprintf(format, args...))
	}
}

func (l *zapGrpcLoggerV2) Warning(args ...interface{}) {
	if l.severity >= warnLevel {
		l.Logger.Warn(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Warningln(args ...interface{}) {
	if l.severity >= warnLevel {
		l.Logger.Warn(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	if l.severity >= warnLevel {
		l.Logger.Warn(fmt.Sprintf(format, args...))
	}
}

func (l *zapGrpcLoggerV2) Error(args ...interface{}) {
	l.Logger.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Errorln(args ...interface{}) {
	l.Logger.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.Logger.Error(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) Fatal(args ...interface{}) {
	l.Logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalln(args ...interface{}) {
	l.Logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) V(level int) bool {
	return level <= l.verbosity
}
