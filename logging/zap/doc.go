// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// ZAP loggers for gRPC.
/*
`grpc_zap` provides server-side interceptors and handler funcitions for using ZAP loggers within gRPC code.

It accepts a user-configured `zap.Logger` object that is:
 - used for logging completed gRPC calls (method called, time elapsed, error code and message, log level)
 - populated into the `context.Context` passed into gRPC handler code.

You can use `Extract` to log into a request-scoped `zap.Logger` instance in your handler code. Moreover you `AddFields`
to the request-scoped `zap.Logger`, that will be propagated for all call depending on the context, including the
interceptor's own "finished RPC" log message.

The latter is very useful when the handler code wants to add additional metadata to the call after extracting it from
the request. For use cases when a "downstream" interceptor needs to log something, please consider using
grpc_commonlog library.

To make sure that ZAP is also receiving the log statements from the gRPC library internals, please call
`ReplaceGrpcLogger`.

Please see examples and tests for examples of use.
*/
package grpc_zap
