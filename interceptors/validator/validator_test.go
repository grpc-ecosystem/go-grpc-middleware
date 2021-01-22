// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package validator_test

import (
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

func TestValidatorTestSuite(t *testing.T) {
	s := &ValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(validator.StreamServerInterceptor()),
				grpc.UnaryInterceptor(validator.UnaryServerInterceptor()),
			},
		},
	}
	suite.Run(t, s)

	cs := &ClientValidatorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(validator.UnaryClientInterceptor()),
			},
		},
	}
	suite.Run(t, cs)
}

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
