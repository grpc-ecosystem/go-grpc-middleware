# Go gRPC Middleware

[![go](https://github.com/grpc-ecosystem/go-grpc-middleware/workflows/go/badge.svg?branch=v2)](https://github.com/grpc-ecosystem/go-grpc-middleware/actions?query=branch%3Av2) [![Go Report Card](https://goreportcard.com/badge/github.com/grpc-ecosystem/go-grpc-middleware)](https://goreportcard.com/report/github.com/grpc-ecosystem/go-grpc-middleware) [![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/v2) [![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE) [![Slack](https://img.shields.io/badge/slack-%23grpc--middleware-brightgreen)](https://gophers.slack.com/archives/CNJL30P4P)

This repository holds [gRPC Go](https://github.com/grpc/grpc-go) Middlewares: interceptors, helpers and utilities.

## Middleware

[gRPC Go](https://github.com/grpc/grpc-go) has support for "interceptors", i.e. [middleware](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81#.gv7tdlghs) that is executed either on the gRPC Server before the request is passed onto the user's application logic, or on the gRPC client either around the user call. It is a perfect way to implement common patterns: auth, logging, tracing, metrics, validation, retries, rate limiting and more, which can be a great generic building blocks that make it easy to build multiple microservices easily.

Especially for observability signals (logging, tracing, metrics) interceptors offers semi-auto-instrumentation that improves consistency of your observability and allows great correlation techniques (e.g. exemplars and trace ID in logs). Demo-ed in [examples](examples).

This repository offers ready-to-use middlewares that implements gRPC interceptors with examples. In some cases dedicated projects offer great interceptors, so this repository skips those, and we link them in the [interceptors](#interceptors) list.

> NOTE: Some middlewares are quite simple to write, so feel free to use this repo as template if you need. It's ok to copy some simpler interceptors if you need more flexibility. This repo can't support all the edge cases you might have.

Additional great feature of interceptors is the fact we can chain those. For example below you can find example server side chain of interceptors with full observabiliy correlation, auth and panic recovery:

```go mdox-exec="sed -n '116,132p' examples/server/main.go"
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
			otelgrpc.UnaryServerInterceptor(),
			srvMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.UnaryServerInterceptor(kit.InterceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(authFn), selector.MatchFunc(allButHealthZ)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			otelgrpc.StreamServerInterceptor(),
			srvMetrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.StreamServerInterceptor(kit.InterceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			selector.StreamServerInterceptor(auth.StreamServerInterceptor(authFn), selector.MatchFunc(allButHealthZ)),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	)
```

This pattern offers clean and explicit shared functionality for all your gRPC methods. Full, buildable examples can be found in [examples](examples) directory.

## Interceptors

This list covers known interceptors that users use for their Go microservices (both in this repo and external). Click on each to see extended examples in `examples_test.go` (also available in [pkg.go.dev](https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/v2))

All paths should work with `go get <path>`.

#### Auth
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth`](interceptors/auth) - a customizable (via `AuthFunc`) piece of auth middleware.

#### Observability
* Metrics:
  * [`github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus`⚡](providers/prometheus) - Prometheus client-side and server-side monitoring middleware. Supports exemplars. Moved from deprecated now [`go-grpc-prometheus`](https://github.com/grpc-ecosystem/go-grpc-prometheus).
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging`](interceptors/logging) - a customizable logging middleware offering extended per request logging. It requires logging provider. Available ones:
  * [`github.com/grpc-ecosystem/go-grpc-middleware/providers/logr`](providers/logr) - adapter for [logr](https://github.com/go-logr/logr).
  * [`github.com/grpc-ecosystem/go-grpc-middleware/providers/logrus`](providers/logrus) - adapter for [logrus](https://github.com/sirupsen/logrus).
  * [`github.com/grpc-ecosystem/go-grpc-middleware/providers/kit`](providers/kit) - adapter for [go-kit/log](https://github.com/go-kit/log).
  * [`github.com/grpc-ecosystem/go-grpc-middleware/providers/phuslog`](providers/phuslog) - adapter for [phuslog](https://github.com/phuslu/log)
  * [`github.com/grpc-ecosystem/go-grpc-middleware/providers/zap`](providers/zap) - adapter for [zap](https://github.com/uber-go/zap).
  * [`github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog`](providers/zerolog) - adapter for [zerolog](https://github.com/rs/zerolog).
* Tracing:
  * (external) [`go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc`](https://go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc) - official OpenTelemetry tracing interceptors as used in [example](examples).
  * (external) [`github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing`](https://pkg.go.dev/github.com/grpc-ecosystem/go-grpc-middleware@v1.4.0/tracing/opentracing) - deprecated [OpenTracing](http://opentracing.io/) client-side and server-side interceptors if you still need it!

#### Client
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry`](interceptors/retry) - a generic gRPC response code retry mechanism, client-side middleware
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout`](interceptors/timeout) - a generic gRPC request timeout, client-side middleware

#### Server
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator`](interceptors/validator) - codegen inbound message validation from `.proto` options.
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery`](interceptors/recovery) - turn panics into gRPC errors (make sure to use those as "last" interceptor, so panic does not skip other interceptors).
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit`](interceptors/ratelimit) - grpc rate limiting by your own limiter.

#### Filtering Interceptor
* [`github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector`](interceptors/selector) - allow users to select given one or more interceptors in certain condition like matching service method.

## Prerequisites

- **[Go](https://golang.org)**: Any one of the **three latest major** [releases](https://golang.org/doc/devel/release.html) are supported.

## Structure of this repository

The main interceptors are available in the subdirectories of the [`interceptors` directory](interceptors) e.g. [`interceptors/validator`](interceptors/validator). Some interceptors work as a standalone piece of middleware with extra parameters e.g. [`auth`](interceptors/auth). Some, like [`interceptors/logging`](interceptors/logging) benefit from adapters maintained as a *separate Go Modules* in [`providers`](providers) directory. For example, there is an adapter for `go-kit/log` logger in [`providers/kit`](providers/kit) directory that works with logging e.g. `logging.StreamServerInterceptor(kit.InterceptorLogger(rpcLogger))`. The separate module, might be a little bit of the initial pain to discover and version in your `go.mod`, but it allows core interceptors to be ultra slim in terms of dependencies.

Some providers are standalone interceptors too. For example [`providers/prometheus`](providers/prometheus) offer metrics middleware (there is no "interceptor/metrics" at the moment).

The [`interceptors` directory](interceptors) also holds generic interceptors that accepts [`Reporter`](interceptors/reporter.go) interface which allows creating your own middlewares with ease.

As you might notice this repository contains multiple modules with different versions ([Go Module specifics](https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories)). Refer to [versions.yaml](versions.yaml) for current modules. We have main module of version 2.x.y and providers modules of lower versions. Since main module is v2, it's module path ends with `v2`:

```
go get github.com/grpc-ecosystem/go-grpc-middleware/v2/<package>
```

For providers modules and packages, since they are v1, no version is added to the path e.g.

```
go get github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus
```

## Changes compared to v1

[go-grpc-middleware v1](https://pkg.go.dev/github.com/grpc-ecosystem/go-grpc-middleware) was created near 2015 and became a popular choice for gRPC users. However, many have changed since then. The main changes of v2 compared to v1:

* Multiple, separate Go modules in "providers" e.g. loggers, so interceptors can be extended without the dependency hell to the core framework (e.g. if use go-kit logger, do you really want to import logrus?). This allows greater extensibility.
* `grpc_opentracing` interceptor was removed. This is because tracing instrumentation evolved. OpenTracing is deprecated and OpenTelemetry has now a [superior tracing interceptor](https://go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc).
* `grpc_ctxtags` interceptor was removed. Custom tags can be added to logging fields using `logging.InjectFields`. Proto option to add logging field was clunky in practice and we don't see any use of it nowadays, so it's removed.
* One of the most powerful interceptor was imported from https://github.com/grpc-ecosystem/go-grpc-prometheus (repo is now deprecated). This consolidation allows easier maintenance, easier use and consistent API.
* Chain interceptors was removed, because `grpc` implemented one.
* Moved to the new proto API (google.golang.org/protobuf).
* All "deciders", so functions that decide what to do based on gRPC service name and method (aka "fullMethodName") now use typed ["interceptors.CallMeta"](interceptors/callmeta.go) allowing better context and explicitness.
* No more snake case package names. We have now single word meaningful package names. If you have collision in package names we recommend adding grpc prefix e.g. `grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"`.
* All the options (if any) are in the form of `<package_name>.With<Option Name>`, with extensibility to add more of them.
* `v2` is the main (default) development branch.

## For Maintainers: Release Process

This assumes we want to release minor version of any module:

1. Understand what has been change and what groups within [versions](versions.yaml) has to be updated.
2. Update group version on v2 branch accordingly.
3. Create new tag for *each module* that has to be released. For the main module `github.com/grpc-ecosystem/go-grpc-middleware/v2` the tag has no prefix (e.g. v2.20.1). For providers (sub modules), the tag version has to have form e.g. `providers/<provider/v1.2.3`. See https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories for details.
4. Once all tags are pushed, draft and create release on GitHub page, mentioning all changed tags in the title. Use auto-generation of notes and remove those that are not relevant for users (e.g. fixing docs).

## License

`go-grpc-middleware` is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
