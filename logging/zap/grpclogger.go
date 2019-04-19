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
func ReplaceGrpcLogger(logger *zap.Logger) {
	zgl := &zapGrpcLogger{logger.With(SystemField, zap.Bool("grpc_log", true))}
	grpclog.SetLoggerV2(zgl)
}

type zapGrpcLogger struct {
	logger *zap.Logger
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (l *zapGrpcLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (l *zapGrpcLogger) Infoln(args ...interface{}) {
	l.logger.Info(fmt.Sprintln(args...))
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (l *zapGrpcLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (l *zapGrpcLogger) Warning(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (l *zapGrpcLogger) Warningln(args ...interface{}) {
	l.logger.Warn(fmt.Sprintln(args...))
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l *zapGrpcLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (l *zapGrpcLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (l *zapGrpcLogger) Errorln(args ...interface{}) {
	l.logger.Error(fmt.Sprintln(args...))
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (l *zapGrpcLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *zapGrpcLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *zapGrpcLogger) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprintln(args...))
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *zapGrpcLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

// V reports whether verbosity level l is at least the requested verbose level.
func (l *zapGrpcLogger) V(_ int) bool {
	// Return true for all levels
	return true
}
