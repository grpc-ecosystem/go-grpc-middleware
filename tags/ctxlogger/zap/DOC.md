# ctxlogger_zap
`import "github.com/grpc-ecosystem/go-grpc-middleware/tags/ctxlogger/zap"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)
* [Examples](#pkg-examples)

## <a name="pkg-overview">Overview</a>
`grpc_zap` is a gRPC logging middleware backed by ZAP loggers

It accepts a user-configured `zap.Logger` that will be used for logging completed gRPC calls. The same `zap.Logger` will
be used for logging completed gRPC calls, and be populated into the `context.Context` passed into gRPC handler code.

On calling `StreamServerInterceptor` or `UnaryServerInterceptor` this logging middleware will add gRPC call information
to the ctx so that it will be present on subsequent use of the `ctxlogger_zap` logger.

You can use `ctxlogger_zap.Extract` to log into a request-scoped `zap.Logger` instance in your handler code.
The fields set on the logger correspond to the grpc_ctxtags.Tags attached to the context.

As `ctxlogger_zap.Extract` will iterate all tags on from `grpc_ctxtags` it is therefore expensive so it is advised that you
extract once at the start of the function from the context and reuse it for the remainder of the function (see examples).

This package also implements request and response *payload* logging, both for server-side and client-side. These will be
logged as structured `jsonbp` fields for every message received/sent (both unary and streaming). For that please use
`Payload*Interceptor` functions for that. Please note that the user-provided function that determines whetether to log
the full request/response payload needs to be written with care, this can significantly slow down gRPC.

ZAP can also be made as a backend for gRPC library internals. For that use `ReplaceGrpcLogger`.

Please see examples and tests for examples of use.

#### Example:

<details>
<summary>Click to expand code.</summary>

```go
x := func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
    // Add fields the ctxtags of the request which will be added to all extracted loggers.
    grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)

    // Extract a single request-scoped zap.Logger and log messages.
    l := ctxlogger_zap.Extract(ctx)
    l.Info("some ping")
    l.Info("another ping")
    return &pb_testproto.PingResponse{Value: ping.Value}, nil
}
return x
```

</details>

#### Example:

<details>
<summary>Click to expand code.</summary>

```go
// Create a server, make sure we put the grpc_ctxtags context before everything else.
server := grpc.NewServer(
    grpc_middleware.WithUnaryServerChain(
        grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
        ctxlogger_zap.UnaryServerInterceptor(zapLogger),
    ),
    grpc_middleware.WithStreamServerChain(
        grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
        ctxlogger_zap.StreamServerInterceptor(zapLogger),
    ),
)
return server
```

</details>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/grpc-ecosystem/go-grpc-middleware](./../../..)
- [github.com/grpc-ecosystem/go-grpc-middleware/tags](./../..)
- [go.uber.org/zap](https://godoc.org/go.uber.org/zap)
- [go.uber.org/zap/zapcore](https://godoc.org/go.uber.org/zap/zapcore)
- [golang.org/x/net/context](https://godoc.org/golang.org/x/net/context)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)

## <a name="pkg-index">Index</a>
* [func AddFields(ctx context.Context, fields ...zapcore.Field)](#AddFields)
* [func Extract(ctx context.Context) \*zap.Logger](#Extract)
* [func StreamServerInterceptor(logger \*zap.Logger) grpc.StreamServerInterceptor](#StreamServerInterceptor)
* [func TagsToFields(ctx context.Context) []zapcore.Field](#TagsToFields)
* [func ToContext(ctx context.Context, logger \*zap.Logger) context.Context](#ToContext)
* [func UnaryServerInterceptor(logger \*zap.Logger) grpc.UnaryServerInterceptor](#UnaryServerInterceptor)

#### <a name="pkg-examples">Examples</a>
* [Package (HandlerUsageUnaryPing)](#example__handlerUsageUnaryPing)
* [Package (Initialization)](#example__initialization)

#### <a name="pkg-files">Package files</a>
[context.go](./context.go) [doc.go](./doc.go) [server_interceptors.go](./server_interceptors.go) 

## <a name="AddFields">func</a> [AddFields](./context.go#L23)
``` go
func AddFields(ctx context.Context, fields ...zapcore.Field)
```
AddFields adds zap fields to the logger.

## <a name="Extract">func</a> [Extract](./context.go#L35)
``` go
func Extract(ctx context.Context) *zap.Logger
```
Extract takes the call-scoped Logger from grpc_zap middleware.

It always returns a Logger that has all the grpc_ctxtags updated.

## <a name="StreamServerInterceptor">func</a> [StreamServerInterceptor](./server_interceptors.go#L19)
``` go
func StreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor that adds zap.Logger to the context.

## <a name="TagsToFields">func</a> [TagsToFields](./context.go#L47)
``` go
func TagsToFields(ctx context.Context) []zapcore.Field
```

## <a name="ToContext">func</a> [ToContext](./context.go#L56)
``` go
func ToContext(ctx context.Context, logger *zap.Logger) context.Context
```

## <a name="UnaryServerInterceptor">func</a> [UnaryServerInterceptor](./server_interceptors.go#L11)
``` go
func UnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)