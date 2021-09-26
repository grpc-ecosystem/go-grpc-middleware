---
layout: default
title: Go gRPC Middleware
nav_order: 0
description: 'Documentation site for the Go gRPC Middleware'
permalink: /
---

# Go gRPC Middleware
{: .fs-9 }

Go gRPC Middleware is a collection of gRPC middleware packages: interceptors, helpers and tools
{: .fs-6 .fw-300 }

[Get started](#getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 } [View it on GitHub](https://github.com/grpc-ecosystem/go-grpc-middleware){: .btn .fs-5 .mb-4 .mb-md-0 }

## Getting started

<a href="https://travis-ci.org/grpc-ecosystem/go-grpc-middleware"><img src="https://img.shields.io/travis/grpc-ecosystem/go-grpc-middleware?logo=travis&logoColor=ffffff&style=flat-square"/></a>
<a href="https://goreportcard.com/report/github.com/grpc-ecosystem/go-grpc-middleware"><img src="https://goreportcard.com/badge/github.com/grpc-ecosystem/go-grpc-middleware?style=flat-square"/></a>
<img src="http://img.shields.io/badge/Godoc-Reference-blue?logoColor=ffffff&style=flat-square"/>
<a href="https://sourcegraph.com/github.com/grpc-ecosystem/go-grpc-middleware/?badge"><img src="https://sourcegraph.com/github.com/grpc-ecosystem/go-grpc-middleware/-/badge.svg?logoColor=ffffff&style=flat-square"/></a>
<a href="https://codecov.io/gh/grpc-ecosystem/go-grpc-middleware"><img src="https://img.shields.io/codecov/c/github/grpc-ecosystem/go-grpc-middleware?logo=codecov&logoColor=ffffff&style=flat-square"/></a>
<a href="https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/LICENSE"><img src="https://img.shields.io/github/license/grpc-ecosystem/go-grpc-middleware?style=flat-square"/></a>
<a href="#status"><img src="https://img.shields.io/badge/quality-production-orange?logoColor=ffffff&style=flat-square"/></a>
<a href="https://github.com/grpc-ecosystem/go-grpc-middleware/stargazers"><img src="https://img.shields.io/github/stars/grpc-ecosystem/go-grpc-middleware?style=flat-square"/></a>
<a href="https://github.com/grpc-ecosystem/go-grpc-middleware/releases"><img src="https://img.shields.io/github/v/release/grpc-ecosystem/go-grpc-middleware?logoColor=ffffff&style=flat-square"/></a>
<a href="https://gophers.slack.com/archives/CNJL30P4P"><img src="https://img.shields.io/badge/slack-grpc--gateway-379c9c?logo=slack&logoColor=ffffff&style=flat-square"/></a>

## Middleware

[gRPC Go](https://github.com/grpc/grpc-go) recently acquired support for
Interceptors, i.e. [middleware](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81#.gv7tdlghs)
that is executed either on the gRPC Server before the request is passed onto the user's application logic, or on the gRPC client around the user call. It is a perfect way to implement common patterns: auth, logging, message, validation, retries, or monitoring.

These are generic building blocks that make it easy to build multiple microservices easily.The purpose of this repository is to act as a go-to point for such reusable functionality. It contains some of them itself, but also will link to useful external repos.

`grpc_middleware` itself provides support for chaining interceptors, here's an example:

```go
import "github.com/grpc-ecosystem/go-grpc-middleware"

myServer := grpc.NewServer(
    grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
        grpc_recovery.StreamServerInterceptor(),
        grpc_ctxtags.StreamServerInterceptor(),
        grpc_opentracing.StreamServerInterceptor(),
        grpc_prometheus.StreamServerInterceptor,
        grpc_zap.StreamServerInterceptor(zapLogger),
        grpc_auth.StreamServerInterceptor(myAuthFunction),
    )),
    grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
        grpc_recovery.UnaryServerInterceptor(),
        grpc_ctxtags.UnaryServerInterceptor(),
        grpc_opentracing.UnaryServerInterceptor(),
        grpc_prometheus.UnaryServerInterceptor,
        grpc_zap.UnaryServerInterceptor(zapLogger),
        grpc_auth.UnaryServerInterceptor(myAuthFunction),
    )),
)
```

## Interceptors

_Please send a PR to add new interceptors or middleware to this list_

#### Auth

- [`grpc_auth`](auth) - a customizable (via `AuthFunc`) piece of auth middleware

#### Logging

- [`grpc_ctxtags`](tags/) - a library that adds a `Tag` map to context, with data populated from request body
- [`grpc_zap`](logging/zap/) - integration of [zap](https://github.com/uber-go/zap) logging library into gRPC handlers.
- [`grpc_logrus`](logging/logrus/) - integration of [logrus](https://github.com/sirupsen/logrus) logging library into gRPC handlers.
- [`grpc_kit`](logging/kit/) - integration of [go-kit/log](https://github.com/go-kit/log) logging library into gRPC handlers.
- [`grpc_grpc_logsettable`](logging/settable/) - a wrapper around `grpclog.LoggerV2` that allows to replace loggers in runtime (thread-safe).

#### Monitoring

- [`grpc_prometheus`⚡](https://github.com/grpc-ecosystem/go-grpc-prometheus) - Prometheus client-side and server-side monitoring middleware
- [`otgrpc`⚡](https://github.com/grpc-ecosystem/grpc-opentracing/tree/master/go/otgrpc) - [OpenTracing](http://opentracing.io/) client-side and server-side interceptors
- [`grpc_opentracing`](tracing/opentracing) - [OpenTracing](http://opentracing.io/) client-side and server-side interceptors with support for streaming and handler-returned tags

#### Client

- [`grpc_retry`](retry/) - a generic gRPC response code retry mechanism, client-side middleware

#### Server

- [`grpc_validator`](validator/) - codegen inbound message validation from `.proto` options
- [`grpc_recovery`](recovery/) - turn panics into gRPC errors
- [`ratelimit`](ratelimit/) - grpc rate limiting by your own limiter

## Status

This code has been running in _production_ since May 2016 as the basis of the gRPC microservices stack at [Improbable](https://improbable.io).

## Contribution

See [CONTRIBUTING.md](https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/CONTRIBUTING.md).

## License

Go gRPC Middleware is released under the Apache 2.0 License. See [LICENSE](https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/LICENSE) for more details.

### Thank you to the contributors of Go gRPC Middleware

<ul class="list-style-none">
{% for contributor in site.github.contributors %}
<li class="d-inline-block mr-1">
<a href="{{ contributor.html_url }}"><img src="{{ contributor.avatar_url }}" width="32" height="32" alt="{{ contributor.login }}"/></a>
</li>
{% endfor %}
</ul>
