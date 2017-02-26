// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// gRPC middleware logging.
/*
`grpc_logging` is a "mother" package for other specific gRPC logging middleware.



Concrete logging middleware for use in user-code handlers is available in subpackages
 * zap
 * logrus

The functions and methods in this package should only be consumed by gRPC logging middleware and other middlewares that
want to add metadata to the logging context of the request.
*/
package grpc_logging
