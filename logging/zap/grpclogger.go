// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_zap

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
	"os"
	"strconv"
)

// ReplaceGrpcLogger sets the given zap.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
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

func ReplaceGrpcLoggerV2(logger *zap.Logger) {
	zgl := &zapGrpcLoggerV2{}
	verbosity := os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL")
	if v, err := strconv.Atoi(verbosity); err == nil {
		zgl.verbosity = v
	}

	logLevel := os.Getenv("GRPC_GO_LOG_SEVERITY_LEVEL")
	switch logLevel {
	case "", "ERROR", "error": // If env is unset, set level to ERROR.
		zgl.severity = error
	case "WARNING", "warning":
		zgl.severity = warn
	case "INFO", "info":
		zgl.severity = info
	}

	zgl.Logger = logger.With(zap.String("system", "grpc"), zap.Bool("grpc_log", true))
	grpclog.SetLoggerV2(zgl)

}

const (
	error = iota
	warn
	info
)

type zapGrpcLoggerV2 struct {
	*zap.Logger
	verbosity int
	severity  int
}

func (l *zapGrpcLoggerV2) Info(args ...interface{}) {
	if l.severity >= info {
		l.Info(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Infoln(args ...interface{}) {
	if l.severity >= info {
		l.Info(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Infof(format string, args ...interface{}) {
	if l.severity >= info {
		l.Info(fmt.Sprintf(format, args...))
	}
}

func (l *zapGrpcLoggerV2) Warning(args ...interface{}) {
	if l.severity >= warn {
		l.Warn(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Warningln(args ...interface{}) {
	if l.severity >= warn {
		l.Warn(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	if l.severity >= warn {
		l.Warn(fmt.Sprintf(format, args...))
	}
}

func (l *zapGrpcLoggerV2) Error(args ...interface{}) {
	l.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Errorln(args ...interface{}) {
	l.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) Fatal(args ...interface{}) {
	l.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalln(args ...interface{}) {
	l.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLoggerV2) V(level int) bool {
	return level <= l.verbosity
}
