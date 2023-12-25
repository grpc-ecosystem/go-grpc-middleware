// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package protovalidate

import (
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type options struct {
	ignoreMessages []protoreflect.MessageType
}

// An Option lets you add options to protovalidate interceptors using With* funcs.
type Option func(*options)

func evaluateOpts(opts []Option) *options {
	optCopy := &options{}
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// WithIgnoreMessages sets the messages that should be ignored by the validator. Use with
// caution and ensure validation is performed elsewhere.
func WithIgnoreMessages(msgs ...protoreflect.MessageType) Option {
	return func(o *options) {
		o.ignoreMessages = msgs
	}
}

func (o *options) shouldIgnoreMessage(m protoreflect.MessageType) bool {
	return slices.ContainsFunc(o.ignoreMessages, func(t protoreflect.MessageType) bool {
		return m == t
	})
}
