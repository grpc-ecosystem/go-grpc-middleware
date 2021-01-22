package skip_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/skip"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

const (
	keyGRPCType = "skip.grpc_type"
	keyService  = "skip.service"
	keyMethod   = "skip.method"
)

func skipped(ctx context.Context) bool {
	return len(tags.Extract(ctx).Values()) <= 0
}

type skipPingService struct {
	testpb.TestServiceServer
}

func checkMetadata(ctx context.Context, grpcType interceptors.GRPCType, service string, method string) error {
	m, _ := metadata.FromIncomingContext(ctx)
	if typeFromMetadata := m.Get(keyGRPCType)[0]; typeFromMetadata != string(grpcType) {
		return status.Errorf(codes.Internal, fmt.Sprintf("expected grpc type %s, got: %s", grpcType, typeFromMetadata))
	}
	if serviceFromMetadata := m.Get(keyService)[0]; serviceFromMetadata != service {
		return status.Errorf(codes.Internal, fmt.Sprintf("expected service %s, got: %s", service, serviceFromMetadata))
	}
	if methodFromMetadata := m.Get(keyMethod)[0]; methodFromMetadata != method {
		return status.Errorf(codes.Internal, fmt.Sprintf("expected method %s, got: %s", method, methodFromMetadata))
	}
	return nil
}

func (s *skipPingService) Ping(ctx context.Context, _ *testpb.PingRequest) (*testpb.PingResponse, error) {
	err := checkMetadata(ctx, interceptors.Unary, testpb.TestServiceFullName, "Ping")
	if err != nil {
		return nil, err
	}

	if skipped(ctx) {
		return &testpb.PingResponse{Value: "skipped"}, nil
	}

	return &testpb.PingResponse{}, nil
}

func (s *skipPingService) PingList(_ *testpb.PingListRequest, stream testpb.TestService_PingListServer) error {
	err := checkMetadata(stream.Context(), interceptors.ServerStream, testpb.TestServiceFullName, "PingList")
	if err != nil {
		return err
	}

	var out testpb.PingListResponse
	if skipped(stream.Context()) {
		out.Value = "skipped"
	}
	return stream.Send(&out)
}

func filter(ctx context.Context, gRPCType interceptors.GRPCType, service string, method string) bool {
	m, _ := metadata.FromIncomingContext(ctx)
	// Set parameters into metadata
	m.Set(keyGRPCType, string(gRPCType))
	m.Set(keyService, service)
	m.Set(keyMethod, method)

	if v := m.Get("skip"); len(v) > 0 && v[0] == "true" {
		return false
	}
	return true
}

func TestSkipSuite(t *testing.T) {
	s := &SkipSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &skipPingService{&testpb.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(skip.UnaryServerInterceptor(tags.UnaryServerInterceptor(), filter)),
				grpc.StreamInterceptor(skip.StreamServerInterceptor(tags.StreamServerInterceptor(), filter)),
			},
		},
	}
	suite.Run(t, s)
}

type SkipSuite struct {
	*testpb.InterceptorTestSuite
}

func (s *SkipSuite) TestPing() {
	t := s.T()

	testCases := []struct {
		name string
		skip bool
	}{
		{
			name: "skip tags interceptor",
			skip: true,
		},
		{
			name: "do not skip",
			skip: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var m metadata.MD
			if tc.skip {
				m = metadata.New(map[string]string{
					"skip": "true",
				})
			}

			resp, err := s.Client.Ping(metadata.NewOutgoingContext(s.SimpleCtx(), m), testpb.GoodPing)
			require.NoError(t, err)

			var value string
			if tc.skip {
				value = "skipped"
			}
			assert.Equal(t, value, resp.Value)
		})
	}
}

func (s *SkipSuite) TestPingList() {
	t := s.T()

	testCases := []struct {
		name string
		skip bool
	}{
		{
			name: "skip tags interceptor",
			skip: true,
		},
		{
			name: "do not skip",
			skip: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var m metadata.MD
			if tc.skip {
				m = metadata.New(map[string]string{
					"skip": "true",
				})
			}

			stream, err := s.Client.PingList(metadata.NewOutgoingContext(s.SimpleCtx(), m), testpb.GoodPingList)
			require.NoError(t, err)

			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				require.NoError(s.T(), err)

				var value string
				if tc.skip {
					value = "skipped"
				}
				assert.Equal(t, value, resp.Value)
			}
		})
	}
}
