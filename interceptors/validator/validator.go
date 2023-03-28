// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func validate(req interface{}, shouldFailFast bool, level logging.Level, logger logging.Logger) error {
	// shouldFailFast tells validator to immediately stop doing further validation after first validation error.
	if shouldFailFast {
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

	// shouldNotFailFast tells validator to continue doing further validation even if after a validation error.
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
