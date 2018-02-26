// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package ctx_glog

import (
	"fmt"
	"time"

	"github.com/json-iterator/go"
	"google.golang.org/grpc/grpclog"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// severity identifies the sort of log: info, warning etc. It also implements
// the flag.Value interface. The -stderrthreshold flag is of type severity and
// should be modified only through the flag.Value interface. The values match
// the corresponding constants in C++.
type Severity int32 // sync/atomic int32

// These constants identify the log levels in order of increasing severity.
// A message written to a high-severity log file is also written to each
// lower-severity log file.
const (
	InfoLevel Severity = iota
	WarningLevel
	ErrorLevel
	FatalLevel
	DebugLevel
)

func (s Severity) String() string {
	switch s {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarningLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	}

	return "unknown"
}

// Defines the key when adding errors using WithError.
var ErrorKey = "error"

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// An entry is the final or intermediate glog logging entry. It contains all
// the fields passed with WithField{,s}. It's finally logged when Debug, Info,
// Warn, Error, Fatal or Panic is called on it. These objects can be reused and
// passed around as much as you wish to avoid field duplication.
type Entry struct {
	Logger grpclog.LoggerV2

	// Contains all the fields set by the user.
	Data Fields

	// Time at which the log entry was created
	Time time.Time

	// Message passed to Debug, Info, Warn, Error, Fatal or Panic
	Message string

	s Severity
}

func NewEntry(logger grpclog.LoggerV2) *Entry {
	return &Entry{
		Logger: logger,
		// Default is three fields, give a little extra room
		Data: make(Fields, 5),
	}
}

// Returns the string representation from the reader and ultimately the
// formatter.
func (entry *Entry) String() string {
	data := make(Fields, len(entry.Data)+1)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
	}
	if l, ok := data["level"]; ok {
		data["fields.level"] = l
	}
	data["msg"] = entry.Message
	data["level"] = entry.s.String()

	serialized, err := json.Marshal(data)
	if err != nil {
		return err.Error()
	}
	str := string(serialized)
	return str
}

// Add an error as single field (using the key defined in ErrorKey) to the Entry.
func (entry *Entry) WithError(err error) *Entry {
	return entry.WithField(ErrorKey, err)
}

// Add a single field to the Entry.
func (entry *Entry) WithField(key string, value interface{}) *Entry {
	return entry.WithFields(Fields{key: value})
}

// Add a map of fields to the Entry.
func (entry *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(entry.Data)+len(fields))
	for k, v := range entry.Data {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return &Entry{Logger: entry.Logger, Data: data}
}

func (entry *Entry) withMessage(args ...interface{}) *Entry {
	entry.Message = fmt.Sprint(args...)
	return entry
}

func (entry *Entry) withMessagef(format string, args ...interface{}) *Entry {
	entry.Message = fmt.Sprintf(format, args...)
	return entry
}

func (entry *Entry) Fatal(args ...interface{}) {
	entry.s = FatalLevel
	entry.Logger.Fatal(entry.withMessage(args...))
}

func (entry *Entry) Fatalln(args ...interface{}) {
	entry.s = FatalLevel
	entry.Logger.Fatalln(entry.withMessage(args...))
}

func (entry *Entry) Fatalf(format string, args ...interface{}) {
	entry.s = FatalLevel
	entry.Logger.Fatal(entry.withMessagef(format, args...))
}

func (entry *Entry) Error(args ...interface{}) {
	entry.s = ErrorLevel
	entry.Logger.Error(entry.withMessage(args...))
}

func (entry *Entry) Errorln(args ...interface{}) {
	entry.s = ErrorLevel
	entry.Logger.Errorln(entry.withMessage(args...))
}

func (entry *Entry) Errorf(format string, args ...interface{}) {
	entry.s = ErrorLevel
	entry.Logger.Error(entry.withMessagef(format, args...))
}

func (entry *Entry) Warning(args ...interface{}) {
	entry.s = WarningLevel
	entry.Logger.Warning(entry.withMessage(args...))
}

func (entry *Entry) Warningln(args ...interface{}) {
	entry.s = WarningLevel
	entry.Logger.Warningln(entry.withMessage(args...))
}

func (entry *Entry) Warningf(format string, args ...interface{}) {
	entry.s = WarningLevel
	entry.Logger.Warning(entry.withMessagef(format, args...))
}

func (entry *Entry) Info(args ...interface{}) {
	entry.s = InfoLevel
	entry.Logger.Info(entry.withMessage(args...))
}

func (entry *Entry) Infoln(args ...interface{}) {
	entry.s = InfoLevel
	entry.Logger.Infoln(entry.withMessage(args...))
}

func (entry *Entry) Infof(format string, args ...interface{}) {
	entry.s = InfoLevel
	entry.Logger.Info(entry.withMessagef(format, args...))
}

func (entry *Entry) Debug(args ...interface{}) {
	if entry.Logger.V(int(DebugLevel)) {
		entry.s = DebugLevel
		entry.Logger.Info(1, entry.withMessage(args...))
	}
}

func (entry *Entry) Debugln(args ...interface{}) {
	if entry.Logger.V(int(DebugLevel)) {
		entry.s = DebugLevel
		entry.Logger.Infoln(entry.withMessage(args...))
	}
}

func (entry *Entry) Debugf(format string, args ...interface{}) {
	if entry.Logger.V(int(DebugLevel)) {
		entry.s = DebugLevel
		entry.Logger.Info(entry.withMessagef(format, args...))
	}
}
