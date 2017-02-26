// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`grpc_zap` is a gRPC logging middleware backed by ZAP loggers

It accepts a user-configured `zap.Logger` that will be used for logging completed gRPC calls. The same `zap.Logger` will
be used for logging completed gRPC calls, and be populated into the `context.Context` passed into gRPC handler code.

You can use `Extract` to log into a request-scoped `zap.Logger` instance in your handler code. `AddFields` adds new fields
to the request-scoped `zap.Logger`. They will be propagated for all call depending on the context, including the
interceptor's own "finished RPC" log message.

ZAP can also be made as a backend for gRPC library internals. For that use `ReplaceGrpcLogger`.


Please see examples and tests for examples of use.
*/
package grpc_zap
