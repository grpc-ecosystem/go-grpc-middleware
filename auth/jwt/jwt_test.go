package grpc_jwt_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_jwt "github.com/grpc-ecosystem/go-grpc-middleware/auth/jwt"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	commonSecret    = "some_secret"
	goodAuthToken   = buildAuthToken(commonSecret)
	badAuthToken    = buildAuthToken("bad_secret")
	brokenAuthToken = "broken_auth_token"
	claims          = jwt.MapClaims{
		"foo": "bar",
		"nbf": float64(time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix()),
	}
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

func buildAuthToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	authToken, _ := token.SignedString([]byte(secret))
	return authToken
}

func assertClaims(t *testing.T, ctx context.Context) {
	unassertedClaims := ctx.Value(grpc_jwt.DefaultContextKey)
	if assert.IsType(t, &jwt.Token{}, unassertedClaims) {
		assertedClaims := unassertedClaims.(*jwt.Token)
		assert.Equal(t, claims, assertedClaims.Claims, "claims from goodAuthToken must be passed around")
	}
}

type assertingPingService struct {
	pb_testproto.TestServiceServer
	T *testing.T
}

func (s *assertingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	assertClaims(s.T, ctx)
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *assertingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	assertClaims(s.T, stream.Context())
	return s.TestServiceServer.PingList(ping, stream)
}

func ctxWithToken(ctx context.Context, scheme string, token string) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}

func TestJwtTestSuite(t *testing.T) {
	authFunc := grpc_jwt.NewAuthFunc([]byte(commonSecret))
	s := &AuthTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &assertingPingService{&grpc_testing.TestPingService{T: t}, t},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
				grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)),
			},
		},
	}
	suite.Run(t, s)
}

type AuthTestSuite struct {
	*grpc_testing.InterceptorTestSuite
}

func (s *AuthTestSuite) TestUnary_NoAuth() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestUnary_BrokenAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", brokenAuthToken), goodPing)
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestUnary_BadAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", badAuthToken), goodPing)
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestUnary_GoodAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", goodAuthToken), goodPing)
	require.NoError(s.T(), err, "no error must occur")
}

func (s *AuthTestSuite) TestUnary_GoodAuthWithPerRpcCredentials() {
	grpcCreds := oauth.TokenSource{TokenSource: &fakeOAuth2TokenSource{accessToken: goodAuthToken}}
	client := s.NewClient(grpc.WithPerRPCCredentials(grpcCreds))
	_, err := client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "no error must occur")
}

func (s *AuthTestSuite) TestStream_NoAuth() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestStream_BrokenAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "bearer", brokenAuthToken), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestStream_BadAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "bearer", badAuthToken), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	assert.Error(s.T(), err, "there must be an error")
	assert.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestStream_GoodAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "Bearer", goodAuthToken), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	pong, err := stream.Recv()
	require.NoError(s.T(), err, "no error must occur")
	require.NotNil(s.T(), pong, "pong must not be nil")
}

func (s *AuthTestSuite) TestStream_GoodAuthWithPerRpcCredentials() {
	grpcCreds := oauth.TokenSource{TokenSource: &fakeOAuth2TokenSource{accessToken: goodAuthToken}}
	client := s.NewClient(grpc.WithPerRPCCredentials(grpcCreds))
	stream, err := client.PingList(s.SimpleCtx(), goodPing)
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
