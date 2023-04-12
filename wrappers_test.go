// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	someKey struct{}
	other   struct{}
)

func TestWrapServerStream(t *testing.T) {
	ctx := context.WithValue(context.TODO(), someKey, 1)
	fake := &fakeServerStream{ctx: ctx}
	wrapped := WrapServerStream(fake)
	assert.NotNil(t, wrapped.Context().Value(someKey), "values from fake must propagate to wrapper")
	wrapped.WrappedContext = context.WithValue(wrapped.Context(), other, 2)
	assert.NotNil(t, wrapped.Context().Value(other), "values from wrapper must be set")
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx         context.Context
	recvMessage any
	sentMessage any
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}

func (f *fakeServerStream) SendMsg(m any) error {
	if f.sentMessage != nil {
		return status.Errorf(codes.AlreadyExists, "fakeServerStream only takes one message, sorry")
	}
	f.sentMessage = m
	return nil
}

func (f *fakeServerStream) RecvMsg(m any) error {
	if f.recvMessage == nil {
		return status.Errorf(codes.NotFound, "fakeServerStream has no message, sorry")
	}
	return nil
}
