// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

type ValidatorTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *ValidatorTestSuite) TestValidPasses_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	assert.NoError(s.T(), err, "no error expected")
}

func (s *ValidatorTestSuite) TestInvalidErrors_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.BadPing)
	assert.Error(s.T(), err, "no error expected")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestValidPasses_ServerStream() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "no error on stream establishment expected")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(s.T(), err, "no error on messages sent occurred")
	}
}

type ClientValidatorTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *ClientValidatorTestSuite) TestValidPasses_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	assert.NoError(s.T(), err, "no error expected")
}

func (s *ClientValidatorTestSuite) TestInvalidErrors_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.BadPing)
	assert.Error(s.T(), err, "error expected")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestInvalidErrors_ServerStream() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.BadPingList)
	require.NoError(s.T(), err, "no error on stream establishment expected")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "error should be received on first message")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestInvalidErrors_BidiStream() {
	stream, err := s.Client.PingStream(s.SimpleCtx())
	require.NoError(s.T(), err, "no error on stream establishment expected")

	require.NoError(s.T(), stream.Send(testpb.GoodPingStream))
	_, err = stream.Recv()
	assert.NoError(s.T(), err, "receiving a good ping should return a good pong")
	require.NoError(s.T(), stream.Send(testpb.GoodPingStream))
	_, err = stream.Recv()
	assert.NoError(s.T(), err, "receiving a good ping should return a good pong")

	require.NoError(s.T(), stream.Send(testpb.BadPingStream))
	_, err = stream.Recv()
	assert.Error(s.T(), err, "receiving a bad ping should return a bad pong")
	assert.Equal(s.T(), codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")

	err = stream.CloseSend()
	assert.NoError(s.T(), err, "there should be no error closing the stream on send")
}

func TestValidatorTestSuite(t *testing.T) {
	sWithNoArgs := &ValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(validator.StreamServerInterceptor()),
				grpc.UnaryInterceptor(validator.UnaryServerInterceptor()),
			},
		},
	}
	suite.Run(t, sWithNoArgs)

	sWithFailFastArgs := &ValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(validator.StreamServerInterceptor(validator.WithFailFast())),
				grpc.UnaryInterceptor(validator.UnaryServerInterceptor(validator.WithFailFast())),
			},
		},
	}
	suite.Run(t, sWithFailFastArgs)

	var gotErrMsgs []string
	onErr := func(ctx context.Context, err error) {
		gotErrMsgs = append(gotErrMsgs, err.Error())
	}
	sWithOnErrFuncArgs := &ValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(validator.StreamServerInterceptor(validator.WithOnValidationErrCallback(onErr))),
				grpc.UnaryInterceptor(validator.UnaryServerInterceptor(validator.WithOnValidationErrCallback(onErr))),
			},
		},
	}
	suite.Run(t, sWithOnErrFuncArgs)
	require.Equal(t, []string{"cannot sleep for more than 10s", "cannot sleep for more than 10s", "cannot sleep for more than 10s"}, gotErrMsgs)

	gotErrMsgs = gotErrMsgs[:0]
	sAll := &ValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(validator.StreamServerInterceptor(validator.WithFailFast(), validator.WithOnValidationErrCallback(onErr))),
				grpc.UnaryInterceptor(validator.UnaryServerInterceptor(validator.WithFailFast(), validator.WithOnValidationErrCallback(onErr))),
			},
		},
	}
	suite.Run(t, sAll)
	require.Equal(t, []string{"cannot sleep for more than 10s", "cannot sleep for more than 10s", "cannot sleep for more than 10s"}, gotErrMsgs)

	csWithNoArgs := &ClientValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(validator.UnaryClientInterceptor()),
			},
		},
	}
	suite.Run(t, csWithNoArgs)

	csWithFailFastArgs := &ClientValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(validator.UnaryServerInterceptor(validator.WithFailFast())),
			},
		},
	}
	suite.Run(t, csWithFailFastArgs)

	gotErrMsgs = gotErrMsgs[:0]
	csWithOnErrFuncArgs := &ClientValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(validator.UnaryServerInterceptor(validator.WithOnValidationErrCallback(onErr))),
			},
		},
	}
	suite.Run(t, csWithOnErrFuncArgs)
	require.Equal(t, []string{"cannot sleep for more than 10s"}, gotErrMsgs)

	gotErrMsgs = gotErrMsgs[:0]
	csAll := &ClientValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(validator.UnaryClientInterceptor(validator.WithFailFast(), validator.WithOnValidationErrCallback(onErr))),
			},
		},
	}
	suite.Run(t, csAll)
	require.Equal(t, []string{"cannot sleep for more than 10s"}, gotErrMsgs)
}
