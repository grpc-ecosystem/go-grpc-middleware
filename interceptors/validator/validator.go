// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// The validateAller interface at protoc-gen-validate main branch.
// See https://github.com/envoyproxy/protoc-gen-validate/pull/468.
type validateAller interface {
	ValidateAll() error
}

// The validate interface starting with protoc-gen-validate v0.6.0.
// See https://github.com/envoyproxy/protoc-gen-validate/pull/455.
type validator interface {
	Validate(all bool) error
}

// The validate interface prior to protoc-gen-validate v0.6.0.
type validatorLegacy interface {
	Validate() error
}

func log(level logging.Level, logger logging.Logger, msg string) {
	if logger != nil {
		logger.Log(level, msg)
	}
}

func validate(req interface{}, d Decider, l Logger) error {
	isFailFast := bool(d())
	level, logger := l()
	if isFailFast {
		switch v := req.(type) {
		case validateAller:
			if err := v.ValidateAll(); err != nil {
				log(level, logger, err.Error())
				return status.Error(codes.InvalidArgument, err.Error())
			}
		case validator:
			if err := v.Validate(true); err != nil {
				log(level, logger, err.Error())
				return status.Error(codes.InvalidArgument, err.Error())
			}
		case validatorLegacy:
			// Fallback to legacy validator
			if err := v.Validate(); err != nil {
				log(level, logger, err.Error())
				return status.Error(codes.InvalidArgument, err.Error())
			}
		}
		return nil
	}
	switch v := req.(type) {
	case validatorLegacy:
		if err := v.Validate(); err != nil {
			log(level, logger, err.Error())
			return status.Error(codes.InvalidArgument, err.Error())
		}
	case validator:
		if err := v.Validate(false); err != nil {
			log(level, logger, err.Error())
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}

func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validate(req, o.shouldFailFast, o.logger); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := validate(req, o.shouldFailFast, o.logger); err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{
			options:      o,
			ServerStream: stream,
		}

		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	*options
	grpc.ServerStream
}

func (s *recvWrapper) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if err := validate(m, s.shouldFailFast, s.logger); err != nil {
		return err
	}
	return nil
}
