package grpc_auth_test

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildDummyAuthzFunction(expectedScheme string, expectedToken string) func(ctx context.Context, fullMethodName string) (context.Context, error) {
	return func(ctx context.Context, fullMethodName string) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, expectedScheme)
		if err != nil {
			return nil, err
		}
		if token != expectedToken {
			return nil, status.Errorf(codes.PermissionDenied, "buildDummyAuthFunction bad token")
		}
		return context.WithValue(ctx, authedMarker, "marker_exists"), nil
	}
}

func TestAuthzTestSuite(t *testing.T) {
	authzFunc := buildDummyAuthzFunction("bearer", commonAuthToken)
	s := &AuthzTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &assertingPingService{&grpc_testing.TestPingService{T: t}, t},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_auth.StreamServerInterceptorAuthz(authzFunc)),
				grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptorAuthz(authzFunc)),
			},
		},
	}
	suite.Run(t, s)
}

func (s *AuthzTestSuite) TestUnary_PassesAuthz() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", commonAuthToken), goodPing)
	require.NoError(s.T(), err, "no error must occur")
}

type AuthzTestSuite struct {
	*grpc_testing.InterceptorTestSuite
}

func (s *AuthzTestSuite) TestUnary_NoAuthz() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthzTestSuite) TestUnary_BadAuthz() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", "bad_token"), goodPing)
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.PermissionDenied, status.Code(err), "must error with permission denied")
}

func (s *AuthzTestSuite) TestStream_NoAuth() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthzTestSuite) TestStream_BadAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "bearer", "bad_token"), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.PermissionDenied, status.Code(err), "must error with permission denied")
}

func (s *AuthzTestSuite) TestStream_PassesAuthz() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "Bearer", commonAuthToken), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	pong, err := stream.Recv()
	require.NoError(s.T(), err, "no error must occur")
	require.NotNil(s.T(), pong, "pong must not be nil")
}

type authzOverrideTestService struct {
	pb_testproto.TestServiceServer
	T *testing.T
}

func (s *authzOverrideTestService) AuthzFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	assert.NotEmpty(s.T, fullMethodName, "method name of caller is passed around")
	return buildDummyAuthzFunction("bearer", overrideAuthToken)(ctx, fullMethodName)
}

func TestAuthzOverrideTestSuite(t *testing.T) {
	authzFunc := buildDummyAuthzFunction("bearer", commonAuthToken)
	s := &AuthzOverrideTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &authzOverrideTestService{&assertingPingService{&grpc_testing.TestPingService{T: t}, t}, t},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_auth.StreamServerInterceptorAuthz(authzFunc)),
				grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptorAuthz(authzFunc)),
			},
		},
	}
	suite.Run(t, s)
}

type AuthzOverrideTestSuite struct {
	*grpc_testing.InterceptorTestSuite
}

func (s *AuthzOverrideTestSuite) TestUnary_PassesAuthz() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", overrideAuthToken), goodPing)
	require.NoError(s.T(), err, "no error must occur")
}

func (s *AuthzOverrideTestSuite) TestStream_PassesAuthz() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "Bearer", overrideAuthToken), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	pong, err := stream.Recv()
	require.NoError(s.T(), err, "no error must occur")
	require.NotNil(s.T(), pong, "pong must not be nil")
}
