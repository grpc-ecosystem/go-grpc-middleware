// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package timeout

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// TimeoutUnaryClientInterceptor returns a new unary client interceptor that sets a timeout on the request context.
func TimeoutUnaryClientInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		timedCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(timedCtx, method, req, reply, cc, opts...)
	}
}
