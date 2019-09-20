// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logrus

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets the given logrus.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLogger(logger *logrus.Entry) {
	grpclog.SetLoggerV2(&logrusGrpcLoggerV2{logger: logger.WithField("system", SystemField)})
}

type logrusGrpcLoggerV2 struct {
	logger *logrus.Entry
}

func (l *logrusGrpcLoggerV2) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *logrusGrpcLoggerV2) Infoln(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *logrusGrpcLoggerV2) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *logrusGrpcLoggerV2) Warning(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *logrusGrpcLoggerV2) Warningln(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *logrusGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *logrusGrpcLoggerV2) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *logrusGrpcLoggerV2) Errorln(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *logrusGrpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *logrusGrpcLoggerV2) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *logrusGrpcLoggerV2) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *logrusGrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *logrusGrpcLoggerV2) V(level int) bool {
	return int(l.logger.Level) >= level
}
