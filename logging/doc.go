// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

//
/*
grpc_logging is a "parent" package for gRPC logging middlewares

General functionality of all middleware

All logging middleware have an `Extract(ctx)` function that provides a request-scoped logger with gRPC-related fields
(service and method names). Additionally, in case a `WithFieldExtractor` is used, the logger will have fields extracted
from the content of the inbound request (unary and server-side stream).

All logging middleware will emit a final log statement. It is based on the error returned by the handler function,
the gRPC status code, an error (if any) and it will emit at a level controlled via `WithLevels`.

This parent package

This particular package is intended for use by other middleware, logging or otherwise. It contains interfaces that other
logging middlewares *should* implement. This allows code to be shared between different implementations.

The `RequestLogFieldExtractorFunc` signature allows users to customize the extraction of request fields to be used as
log fields in middlewares. There are two implementations: one (default) that relies on optional code-generated
`ExtractLogFields()` methods on protobuf structs, and another that uses tagging.

Implementations

There are two implementations at the moment: logrus and zap

See relevant packages below.
*/
package grpc_logging
