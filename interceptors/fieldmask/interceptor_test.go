// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package fieldmask

import (
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestFieldMaskSuite(t *testing.T) {
	s := &FieldMaskSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &testpb.TestPingService{},
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(
					UnaryServerInterceptor(DefaultFilterFunc),
				),
			},
		},
	}
	suite.Run(t, s)
}

type FieldMaskSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *FieldMaskSuite) TestUnary_ReturnAllResponse() {
	resp, err := s.Client.Ping(s.SimpleCtx(), &testpb.PingRequest{Value: "1"})
	assert.Equal(s.T(), nil, err)
	expected := &testpb.PingResponse{
		Value: "1",
	}
	assert.Equal(s.T(), expected.Counter, resp.Counter)
	assert.Equal(s.T(), expected.Value, resp.Value)
}

func (s *FieldMaskSuite) TestUnary_NoReturnValueResponse() {
	resp, err := s.Client.Ping(s.SimpleCtx(), &testpb.PingRequest{Value: "1", FieldMask: &fieldmaskpb.FieldMask{
		Paths: []string{"counter"},
	}})
	assert.Equal(s.T(), nil, err)
	expected := &testpb.PingResponse{
		Value: "",
	}
	assert.Equal(s.T(), expected.Counter, resp.Counter)
	assert.Equal(s.T(), expected.Value, resp.Value)
}
