// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package validator_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ValidatorTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *ValidatorTestSuite) TestValidPasses_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	s.Assert().NoError(err, "no error expected")
}

func (s *ValidatorTestSuite) TestInvalidErrors_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.BadPing)
	s.Require().Error(err, "no error expected")
	s.Assert().Equal(codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestValidPasses_ServerStream() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	s.Require().NoError(err, "no error on stream establishment expected")
	for {
		_, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		s.Assert().NoError(err, "no error on messages sent occurred")
	}
}

type ClientValidatorTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *ClientValidatorTestSuite) TestValidPasses_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	s.Assert().NoError(err, "no error expected")
}

func (s *ClientValidatorTestSuite) TestInvalidErrors_Unary() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.BadPing)
	s.Require().Error(err, "error expected")
	s.Assert().Equal(codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestInvalidErrors_ServerStream() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.BadPingList)
	s.Require().NoError(err, "no error on stream establishment expected")
	_, err = stream.Recv()
	s.Require().Error(err, "error should be received on first message")
	s.Assert().Equal(codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")
}

func (s *ValidatorTestSuite) TestInvalidErrors_BidiStream() {
	stream, err := s.Client.PingStream(s.SimpleCtx())
	s.Require().NoError(err, "no error on stream establishment expected")

	s.Require().NoError(stream.Send(testpb.GoodPingStream))
	_, err = stream.Recv()
	s.Assert().NoError(err, "receiving a good ping should return a good pong")
	s.Require().NoError(stream.Send(testpb.GoodPingStream))
	_, err = stream.Recv()
	s.Assert().NoError(err, "receiving a good ping should return a good pong")

	s.Require().NoError(stream.Send(testpb.BadPingStream))
	_, err = stream.Recv()
	s.Require().Error(err, "receiving a bad ping should return a bad pong")
	s.Assert().Equal(codes.InvalidArgument, status.Code(err), "gRPC status must be InvalidArgument")

	err = stream.CloseSend()
	s.Assert().NoError(err, "there should be no error closing the stream on send")
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
