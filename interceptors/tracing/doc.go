// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

/*
tracing is a "parent" package for gRPC logging middlewares.

This middleware relies on OpenTracing as our tracing interface.

OpenTracing Interceptors

These are both client-side and server-side interceptors for OpenTracing. They are a provider-agnostic, with backends
such as Zipkin, or Google Stackdriver Trace.

For a service that sends out requests and receives requests, you *need* to use both, otherwise downstream requests will
not have the appropriate requests propagated.

All server-side spans are tagged with grpc_ctxtags information.

For more information see:
http://opentracing.io/documentation/
https://github.com/opentracing/specification/blob/master/semantic_conventions.md
*/
package tracing
