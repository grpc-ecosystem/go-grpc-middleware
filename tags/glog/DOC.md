# ctx_glog
`import "github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
`ctx_glog` is a ctxlogger that is backed by glog

It accepts a user-configured `Logger` that will be used for logging. The same `Logger` will
be populated into the `context.Context` passed into gRPC handler code.

You can use `ctx_glog.Extract` to log into a request-scoped `Logger` instance in your handler code.

As `ctx_glog.Extract` will iterate all tags on from `grpc_ctxtags` it is therefore expensive so it is advised that you
extract once at the start of the function from the context and reuse it for the remainder of the function (see examples).

Please see examples and tests for examples of use.

## <a name="pkg-imports">Imported Packages</a>

- [github.com/golang/glog](https://godoc.org/github.com/golang/glog)
- [github.com/grpc-ecosystem/go-grpc-middleware/tags](./..)
- [github.com/json-iterator/go](https://godoc.org/github.com/json-iterator/go)
- [golang.org/x/net/context](https://godoc.org/golang.org/x/net/context)
- [google.golang.org/grpc/grpclog](https://godoc.org/google.golang.org/grpc/grpclog)

## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [func AddFields(ctx context.Context, fields Fields)](#AddFields)
* [func ToContext(ctx context.Context, entry \*Entry) context.Context](#ToContext)
* [type Entry](#Entry)
  * [func Extract(ctx context.Context) \*Entry](#Extract)
  * [func NewEntry(logger grpclog.LoggerV2) \*Entry](#NewEntry)
  * [func (entry \*Entry) Debug(args ...interface{})](#Entry.Debug)
  * [func (entry \*Entry) Debugf(format string, args ...interface{})](#Entry.Debugf)
  * [func (entry \*Entry) Debugln(args ...interface{})](#Entry.Debugln)
  * [func (entry \*Entry) Error(args ...interface{})](#Entry.Error)
  * [func (entry \*Entry) Errorf(format string, args ...interface{})](#Entry.Errorf)
  * [func (entry \*Entry) Errorln(args ...interface{})](#Entry.Errorln)
  * [func (entry \*Entry) Fatal(args ...interface{})](#Entry.Fatal)
  * [func (entry \*Entry) Fatalf(format string, args ...interface{})](#Entry.Fatalf)
  * [func (entry \*Entry) Fatalln(args ...interface{})](#Entry.Fatalln)
  * [func (entry \*Entry) Info(args ...interface{})](#Entry.Info)
  * [func (entry \*Entry) Infof(format string, args ...interface{})](#Entry.Infof)
  * [func (entry \*Entry) Infoln(args ...interface{})](#Entry.Infoln)
  * [func (entry \*Entry) String() string](#Entry.String)
  * [func (entry \*Entry) Warning(args ...interface{})](#Entry.Warning)
  * [func (entry \*Entry) Warningf(format string, args ...interface{})](#Entry.Warningf)
  * [func (entry \*Entry) Warningln(args ...interface{})](#Entry.Warningln)
  * [func (entry \*Entry) WithError(err error) \*Entry](#Entry.WithError)
  * [func (entry \*Entry) WithField(key string, value interface{}) \*Entry](#Entry.WithField)
  * [func (entry \*Entry) WithFields(fields Fields) \*Entry](#Entry.WithFields)
* [type Fields](#Fields)
* [type Severity](#Severity)
  * [func (s Severity) String() string](#Severity.String)

#### <a name="pkg-files">Package files</a>
[context.go](./context.go) [doc.go](./doc.go) [glogger.go](./glogger.go) [noop.go](./noop.go) [types.go](./types.go) 

## <a name="pkg-variables">Variables</a>
``` go
var ErrorKey = "error"
```
Defines the key when adding errors using WithError.

``` go
var Logger grpclog.LoggerV2 = &glogger{}
```

## <a name="AddFields">func</a> [AddFields](./context.go#L20)
``` go
func AddFields(ctx context.Context, fields Fields)
```
AddFields adds glog fields to the logger.

## <a name="ToContext">func</a> [ToContext](./context.go#L58)
``` go
func ToContext(ctx context.Context, entry *Entry) context.Context
```
ToContext adds the Entry to the context for extraction later.
Returning the new context that has been created.

## <a name="Entry">type</a> [Entry](./types.go#L59-L72)
``` go
type Entry struct {
    Logger grpclog.LoggerV2

    // Contains all the fields set by the user.
    Data Fields

    // Time at which the log entry was created
    Time time.Time

    // Message passed to Debug, Info, Warn, Error, Fatal or Panic
    Message string
    // contains filtered or unexported fields
}
```
An entry is the final or intermediate glog logging entry. It contains all
the fields passed with WithField{,s}. It's finally logged when Debug, Info,
Warn, Error, Fatal or Panic is called on it. These objects can be reused and
passed around as much as you wish to avoid field duplication.

### <a name="Extract">func</a> [Extract](./context.go#L34)
``` go
func Extract(ctx context.Context) *Entry
```
Extract takes the call-scoped Entry from ctx_glog middleware.

If the ctx_glog middleware wasn't used, a no-op `Entry` is returned. This makes it safe to
use regardless.

### <a name="NewEntry">func</a> [NewEntry](./types.go#L74)
``` go
func NewEntry(logger grpclog.LoggerV2) *Entry
```

### <a name="Entry.Debug">func</a> (\*Entry) [Debug](./types.go#L205)
``` go
func (entry *Entry) Debug(args ...interface{})
```

### <a name="Entry.Debugf">func</a> (\*Entry) [Debugf](./types.go#L219)
``` go
func (entry *Entry) Debugf(format string, args ...interface{})
```

### <a name="Entry.Debugln">func</a> (\*Entry) [Debugln](./types.go#L212)
``` go
func (entry *Entry) Debugln(args ...interface{})
```

### <a name="Entry.Error">func</a> (\*Entry) [Error](./types.go#L160)
``` go
func (entry *Entry) Error(args ...interface{})
```

### <a name="Entry.Errorf">func</a> (\*Entry) [Errorf](./types.go#L170)
``` go
func (entry *Entry) Errorf(format string, args ...interface{})
```

### <a name="Entry.Errorln">func</a> (\*Entry) [Errorln](./types.go#L165)
``` go
func (entry *Entry) Errorln(args ...interface{})
```

### <a name="Entry.Fatal">func</a> (\*Entry) [Fatal](./types.go#L145)
``` go
func (entry *Entry) Fatal(args ...interface{})
```

### <a name="Entry.Fatalf">func</a> (\*Entry) [Fatalf](./types.go#L155)
``` go
func (entry *Entry) Fatalf(format string, args ...interface{})
```

### <a name="Entry.Fatalln">func</a> (\*Entry) [Fatalln](./types.go#L150)
``` go
func (entry *Entry) Fatalln(args ...interface{})
```

### <a name="Entry.Info">func</a> (\*Entry) [Info](./types.go#L190)
``` go
func (entry *Entry) Info(args ...interface{})
```

### <a name="Entry.Infof">func</a> (\*Entry) [Infof](./types.go#L200)
``` go
func (entry *Entry) Infof(format string, args ...interface{})
```

### <a name="Entry.Infoln">func</a> (\*Entry) [Infoln](./types.go#L195)
``` go
func (entry *Entry) Infoln(args ...interface{})
```

### <a name="Entry.String">func</a> (\*Entry) [String](./types.go#L84)
``` go
func (entry *Entry) String() string
```
Returns the string representation from the reader and ultimately the
formatter.

### <a name="Entry.Warning">func</a> (\*Entry) [Warning](./types.go#L175)
``` go
func (entry *Entry) Warning(args ...interface{})
```

### <a name="Entry.Warningf">func</a> (\*Entry) [Warningf](./types.go#L185)
``` go
func (entry *Entry) Warningf(format string, args ...interface{})
```

### <a name="Entry.Warningln">func</a> (\*Entry) [Warningln](./types.go#L180)
``` go
func (entry *Entry) Warningln(args ...interface{})
```

### <a name="Entry.WithError">func</a> (\*Entry) [WithError](./types.go#L114)
``` go
func (entry *Entry) WithError(err error) *Entry
```
Add an error as single field (using the key defined in ErrorKey) to the Entry.

### <a name="Entry.WithField">func</a> (\*Entry) [WithField](./types.go#L119)
``` go
func (entry *Entry) WithField(key string, value interface{}) *Entry
```
Add a single field to the Entry.

### <a name="Entry.WithFields">func</a> (\*Entry) [WithFields](./types.go#L124)
``` go
func (entry *Entry) WithFields(fields Fields) *Entry
```
Add a map of fields to the Entry.

## <a name="Fields">type</a> [Fields](./types.go#L53)
``` go
type Fields map[string]interface{}
```
Fields type, used to pass to `WithFields`.

## <a name="Severity">type</a> [Severity](./types.go#L19)
``` go
type Severity int32 // sync/atomic int32

```
severity identifies the sort of log: info, warning etc. It also implements
the flag.Value interface. The -stderrthreshold flag is of type severity and
should be modified only through the flag.Value interface. The values match
the corresponding constants in C++.

``` go
const (
    InfoLevel Severity = iota
    WarningLevel
    ErrorLevel
    FatalLevel
    DebugLevel
)
```
These constants identify the log levels in order of increasing severity.
A message written to a high-severity log file is also written to each
lower-severity log file.

### <a name="Severity.String">func</a> (Severity) [String](./types.go#L32)
``` go
func (s Severity) String() string
```

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)