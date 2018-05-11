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
// ReplaceGrpcLoggerV2 replaces the grpc_log.LoggerV2 with the input zap.Logger
// This logger adheres to the grpc go environment variables GRPC_GO_LOG_VERBOSITY_LEVEL and GRPC_GO_LOG_SEVERITY_LEVEL.
func ReplaceGrpcLoggerV2(logger *zap.Logger) {
	zgl := &zapGrpcLoggerV2{}
	verbosity := os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL")
	if v, err := strconv.Atoi(verbosity); err == nil {
		zgl.verbosity = v
	}

	logLevel := os.Getenv("GRPC_GO_LOG_SEVERITY_LEVEL")
	switch logLevel {
	case "", "ERROR", "error": // If env is unset, set level to ERROR.
		zgl.severity = errorLevel
	case "WARNING", "warning":
		zgl.severity = warnLevel
	case "INFO", "info":
		zgl.severity = infoLevel
	}

	zgl.Logger = logger.With(zap.String("system", "grpc"), zap.Bool("grpc_log", true))
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
		l.Info(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Infoln(args ...interface{}) {
	if l.severity >= infoLevel {
		l.Info(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Infof(format string, args ...interface{}) {
	if l.severity >= infoLevel {
		l.Info(fmt.Sprintf(format, args...))
	}
}

func (l *zapGrpcLoggerV2) Warning(args ...interface{}) {
	if l.severity >= warnLevel {
		l.Warn(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Warningln(args ...interface{}) {
	if l.severity >= warnLevel {
		l.Warn(fmt.Sprint(args...))
	}
}

func (l *zapGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	if l.severity >= warnLevel {
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
