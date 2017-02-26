package grpc_logrus

import (
	"github.com/Sirupsen/logrus"
	"io/ioutil"
)

var (
	nullLogger = &logrus.Logger{
		Out:       ioutil.Discard,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.PanicLevel,
	}
)
