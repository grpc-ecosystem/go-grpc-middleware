// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// gRPC middleware logging.
/*
`grpc_logging` is a "mother" package for other specific gRPC logging middleware.

General functionality across all logging middleware:
 * Extract(ctx) function that provides a request-scoped logger with pre-defined fields
 * log statement on completion of handling with customizeable log levels, gRPC status code and error message logging
 * automatic request field to log field extraction, either through code-generated data or field annotations

Concrete logging middleware for use in user-code handlers is available in subpackages
 * zap
 * logrus

The functions and methods in this package should only be consumed by gRPC logging middleware and other middlewares that
want to add metadata to the logging context of the request.
*/
package grpc_logging
