// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/util/metautils"
)

var authedMarker struct{}

var (
	commonAuthToken   = "some_good_token"
	overrideAuthToken = "override_token"
)

// TODO(mwitkow): Add auth from metadata client dialer, which requires TLS.

func buildDummyAuthFunction(expectedScheme string, expectedToken string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, expectedScheme)
		if err != nil {
			return nil, err
		}
		if token != expectedToken {
			return nil, status.Errorf(codes.PermissionDenied, "buildDummyAuthFunction bad token")
		}
		return context.WithValue(ctx, authedMarker, "marker_exists"), nil
	}
}

func assertAuthMarkerExists(t *testing.T, ctx context.Context) {
	assert.Equal(t, "marker_exists", ctx.Value(authedMarker).(string), "auth marker from buildDummyAuthFunction must be passed around")
}

type assertingPingService struct {
	testpb.TestServiceServer
	T *testing.T
}

func (s *assertingPingService) PingError(ctx context.Context, ping *testpb.PingErrorRequest) (*testpb.PingErrorResponse, error) {
	assertAuthMarkerExists(s.T, ctx)
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *assertingPingService) PingList(ping *testpb.PingListRequest, stream testpb.TestService_PingListServer) error {
	assertAuthMarkerExists(s.T, stream.Context())
	return s.TestServiceServer.PingList(ping, stream)
}

func ctxWithToken(ctx context.Context, scheme string, token string) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}

func TestAuthTestSuite(t *testing.T) {
	authFunc := buildDummyAuthFunction("bearer", commonAuthToken)
	s := &AuthTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &assertingPingService{&testpb.TestPingService{T: t}, t},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(auth.StreamServerInterceptor(authFunc)),
				grpc.UnaryInterceptor(auth.UnaryServerInterceptor(authFunc)),
			},
		},
	}
	suite.Run(t, s)
}

type AuthTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *AuthTestSuite) TestUnary_NoAuth() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestUnary_BadAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", "bad_token"), testpb.GoodPing)
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.PermissionDenied, status.Code(err), "must error with permission denied")
}

func (s *AuthTestSuite) TestUnary_PassesAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", commonAuthToken), testpb.GoodPing)
	require.NoError(s.T(), err, "no error must occur")
}

func (s *AuthTestSuite) TestUnary_PassesWithPerRpcCredentials() {
	grpcCreds := oauth.TokenSource{TokenSource: &fakeOAuth2TokenSource{accessToken: commonAuthToken}}
	client := s.NewClient(grpc.WithPerRPCCredentials(grpcCreds))
	_, err := client.Ping(s.SimpleCtx(), testpb.GoodPing)
	require.NoError(s.T(), err, "no error must occur")
}

func (s *AuthTestSuite) TestStream_NoAuth() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestStream_BadAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "bearer", "bad_token"), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.PermissionDenied, status.Code(err), "must error with permission denied")
}

func (s *AuthTestSuite) TestStream_PassesAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "Bearer", commonAuthToken), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	pong, err := stream.Recv()
	require.NoError(s.T(), err, "no error must occur")
	require.NotNil(s.T(), pong, "pong must not be nil")
}

func (s *AuthTestSuite) TestStream_PassesWithPerRpcCredentials() {
	grpcCreds := oauth.TokenSource{TokenSource: &fakeOAuth2TokenSource{accessToken: commonAuthToken}}
	client := s.NewClient(grpc.WithPerRPCCredentials(grpcCreds))
	stream, err := client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	pong, err := stream.Recv()
	require.NoError(s.T(), err, "no error must occur")
	require.NotNil(s.T(), pong, "pong must not be nil")
}

type authOverrideTestService struct {
	testpb.TestServiceServer
	T *testing.T
}

func (s *authOverrideTestService) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	assert.NotEmpty(s.T, fullMethodName, "method name of caller is passed around")
	return buildDummyAuthFunction("bearer", overrideAuthToken)(ctx)
}

func TestAuthOverrideTestSuite(t *testing.T) {
	authFunc := buildDummyAuthFunction("bearer", commonAuthToken)
	s := &AuthOverrideTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &authOverrideTestService{&assertingPingService{&testpb.TestPingService{T: t}, t}, t},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(auth.StreamServerInterceptor(authFunc)),
				grpc.UnaryInterceptor(auth.UnaryServerInterceptor(authFunc)),
			},
		},
	}
	suite.Run(t, s)
}

type AuthOverrideTestSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *AuthOverrideTestSuite) TestUnary_PassesAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", overrideAuthToken), testpb.GoodPing)
	require.NoError(s.T(), err, "no error must occur")
}

func (s *AuthOverrideTestSuite) TestStream_PassesAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "Bearer", overrideAuthToken), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	pong, err := stream.Recv()
	require.NoError(s.T(), err, "no error must occur")
	require.NotNil(s.T(), pong, "pong must not be nil")
}

// fakeOAuth2TokenSource implements a fake oauth2.TokenSource for the purpose of credentials test.
type fakeOAuth2TokenSource struct {
	accessToken string
}

func (ts *fakeOAuth2TokenSource) Token() (*oauth2.Token, error) {
	t := &oauth2.Token{
		AccessToken: ts.accessToken,
		Expiry:      time.Now().Add(1 * time.Minute),
		TokenType:   "bearer",
	}
	return t, nil
}
