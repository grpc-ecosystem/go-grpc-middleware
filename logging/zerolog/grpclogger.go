// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package zerolog

import (
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLoggerV2 sets the given zerolog.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLoggerV2(logger *zerolog.Logger) {
	grpclog.SetLoggerV2(&zerologGrpcLoggerV2{
		logger: logger,
	})
}

type zerologGrpcLoggerV2 struct {
	logger *zerolog.Logger
}

func (l *zerologGrpcLoggerV2) Info(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Infoln(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Infof(format string, args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprintf(format, args...))
}

func (l *zerologGrpcLoggerV2) Warning(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Warningln(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Warningf(format string, args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprintf(format, args...))
}

func (l *zerologGrpcLoggerV2) Error(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Errorln(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprintf(format, args...))
}

func (l *zerologGrpcLoggerV2) Fatal(args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Fatalln(args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprint(args...))
}

func (l *zerologGrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msg(fmt.Sprintf(format, args...))
}

func (l *zerologGrpcLoggerV2) V(level int) bool {
	return zerolog.Level(level) <= l.logger.GetLevel()
}
