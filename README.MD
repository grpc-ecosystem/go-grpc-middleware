# Go gRPC Middleware - Interceptor Helpers

[![Travis Build](https://travis-ci.org/mwitkow/go-grpc-middleware.svg)](https://travis-ci.org/mwitkow/go-grpc-middleware)
[![Go Report Card](http://goreportcard.com/badge/mwitkow/go-flagz)](http://goreportcard.com/report/mwitkow/go-grpc-middleware)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/mwitkow/go-grpc-middleware)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

[gRPC Go](https://github.com/grpc/grpc-go) Interceptor [middleware](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81#.gv7tdlghs). Chain them, wrap them... arrange
them in any pattern that normal `http.Handler` middleware would do :)

## Interceptors

[gRPC Go](https://github.com/grpc/grpc-go) recently acquired support for Interceptors, i.e. middleware that is executed
by a gRPC Server before the request is passed onto the user's application logic. It is a perfect way to implement
common patters: auth, logging, or monitoring (e.g. [Prometheus](https://github.com/mwitkow/go-grpc-prometheus))

Unfortunately out of the box, grpc-go allows you to have only one interceptor, which means you need
to implement multiple interceptors manually... every time.

This repo contains tooling for gRPC interceptors, allowing you to achieve what many people do with
HTTP middleware: auth, logging, pre-processing etc.

## Chaining

Simple way of turning a multiple interceptors into a single interceptor. E.g.

```go
import "github.com/mwitkow/go-grpc-middleware"

myServer := grpc.NewServer(
    grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(loggingStream, monitoringStream, authStream)),
    grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(loggingUnary, monitoringUnary, authUnary),
)
```

or more simply when initializing a server:

```go
myServer := grpc.NewServer(
    grpc_middleware.WithStreamServerChain(loggingStream, monitoringStream, authStream),
    grpc_middleware.WithUnaryServerChain(loggingUnary, monitoringUnary, authUnary),
)
```

These interceptors will be executed from left to right: logging, monitoring and auth.

## Context propagation

It is incredibly useful to pass values around from one interceptor to another, similarly to HTTP
Middleware design. For example, you may want to pass the identity of the caller from the 
auth interceptor all the way to the calling function.

### Unary (1:1) RPCs 

This is relatively simple, since `context.Context` is appendable, for example:

```go
func FakeAuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
   newCtx := context.WithValue(ctx, "user_id", "john@example.com")
   return handler(newCtx, req)
}
```

### Streaming (N:1, 1:M, N:M) RPCs

Unfortunately, it's not as easy for streaming RPCs. These have the `context.Context` embedded within
the `grpc.ServerStream` object. To pass values through context, a wrapper (`WrappedServerStream`) is
needed. For example:

```go
func FakeAuthStreamingInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
   newStream := grpc_middleware.WrapServerStream(stream)
   newStream.WrappedContext = context.WithValue(ctx, "user_id", "john@example.com")
   return handler(srv, stream)
}
```


## Status

This code has been inspired by the gRPC interceptor design discussion, and the [upstream PR](https://github.com/grpc/grpc-go/pull/653).

It is well-tested but still experimental. Additional tooling will be added, and contributions are 
welcome.

## License

`go-grpc-middleware` is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
