// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate_test

import (
	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

type ValidatorTestSuite struct {
	*testvalidate.InterceptorTestSuite
}

func (s *ValidatorTestSuite) TestValidEmail_Unary() {
	_, err := s.Client.Send(s.SimpleCtx(), testvalidate.GoodUnaryRequest)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), codes.OK, status.Code(err), "gRPC status must be OK")
}

func (s *ValidatorTestSuite) TestInvalidEmail_Unary() {
	_, err := s.Client.Send(s.SimpleCtx(), testvalidate.BadUnaryRequest)
	require.Error(s.T(), err)
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestValidEmail_ServerStream() {
	stream, err := s.Client.SendStream(s.SimpleCtx(), testvalidate.GoodStreamRequest)
	require.NoError(s.T(), err, "no error on stream establishment expected")

	_, err = stream.Recv()
	assert.NoError(s.T(), err, "error should be received on first message")
	assert.Equal(s.T(), codes.OK, status.Code(err), "gRPC status must be OK")
}

func (s *ValidatorTestSuite) TestInvalidEmail_Stream() {
	stream, err := s.Client.SendStream(s.SimpleCtx(), testvalidate.BadStreamRequest)
	require.NoError(s.T(), err)

	_, err = stream.Recv()
	assert.Error(s.T(), err, "error should be received on first message")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func TestValidatorTestSuite(t *testing.T) {
	validator, err := protovalidate.New()
	require.NoError(t, err)

	sWithNoArgs := &ValidatorTestSuite{
		InterceptorTestSuite: &testvalidate.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(protovalidate_middleware.StreamServerInterceptor(validator)),
				grpc.UnaryInterceptor(protovalidate_middleware.UnaryServerInterceptor(validator)),
			},
		},
	}
	suite.Run(t, sWithNoArgs)
}
