// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package protovalidate

import (
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// DefaultErrorConverter returns InvalidArgument status with error message from validator.
func DefaultErrorConverter(err error) error {
	return status.Error(codes.InvalidArgument, err.Error())
}

var (
	defaultOptions = &options{
		errorConverter: DefaultErrorConverter,
	}
)

type options struct {
	ignoreMessages []protoreflect.MessageType
	errorConverter ErrorConverter
}

// An Option lets you add options to protovalidate interceptors using With* funcs.
type Option func(*options)

func evaluateOpts(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
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

// ErrorConverter function customize the error returned by protovalidate.Validator.
type ErrorConverter = func(err error) error

// WithErrorConverter customizes the function for mapping errors.
//
// By default, DefaultErrorConverter used.
func WithErrorConverter(errorConverter ErrorConverter) Option {
	return func(o *options) {
		o.errorConverter = errorConverter
	}
}
