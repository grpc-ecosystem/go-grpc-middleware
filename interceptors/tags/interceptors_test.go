package ctxtags_test

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	grpctesting "github.com/grpc-ecosystem/go-grpc-middleware/grpctesting"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/grpctesting/testproto"
	ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/interceptors/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

var (
	goodPing    = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
	anotherPing = &pb_testproto.PingRequest{Value: "else", SleepTimeMs: 9999}
)

func tagsToJson(value map[string]string) string {
	str, _ := json.Marshal(value)
	return string(str)
}

func tagsFromJson(t *testing.T, jstring string) map[string]string {
	var msgMapTemplate map[string]string
	err := json.Unmarshal([]byte(jstring), &msgMapTemplate)
	if err != nil {
		t.Fatalf("failed unmarshaling tags from response %v", err)
	}
	return msgMapTemplate
}

type tagPingBack struct {
	pb_testproto.TestServiceServer
}

func (s *tagPingBack) Ping(ctx context.Context, _ *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	return &pb_testproto.PingResponse{Value: tagsToJson(ctxtags.Extract(ctx).Values())}, nil
}

func (s *tagPingBack) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *tagPingBack) PingList(_ *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	out := &pb_testproto.PingResponse{Value: tagsToJson(ctxtags.Extract(stream.Context()).Values())}
	return stream.Send(out)
}

func (s *tagPingBack) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

func (s *tagPingBack) PingStream(stream pb_testproto.TestService_PingStreamServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		out := &pb_testproto.PingResponse{Value: tagsToJson(ctxtags.Extract(stream.Context()).Values())}
		err = stream.Send(out)
		if err != nil {
			return err
		}
	}
}
func TestTaggingSuite(t *testing.T) {
	opts := []ctxtags.Option{
		ctxtags.WithFieldExtractor(ctxtags.CodeGenRequestFieldExtractor),
	}
	s := &TaggingSuite{
		InterceptorTestSuite: &grpctesting.InterceptorTestSuite{
			TestService: &tagPingBack{&grpctesting.TestPingService{T: t}},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(ctxtags.StreamServerInterceptor(opts...)),
				grpc.UnaryInterceptor(ctxtags.UnaryServerInterceptor(opts...)),
			},
		},
	}
	suite.Run(t, s)
}

type TaggingSuite struct {
	*grpctesting.InterceptorTestSuite
}

func (s *TaggingSuite) SetupTest() {
}

func (s *TaggingSuite) TestPing_WithCustomTags() {
	resp, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "must not be an error on a successful call")

	tags := tagsFromJson(s.T(), resp.Value)
	assert.Equal(s.T(), "something", tags["grpc.request.value"], "the tags should contain the correct request value")
	assert.Contains(s.T(), tags, "peer.address", "the tags should contain a peer address")
	require.Len(s.T(), tags, 2)
}

func (s *TaggingSuite) TestPing_WithDeadline() {
	ctx, _ := context.WithDeadline(context.TODO(), time.Now().AddDate(0, 0, 5))
	resp, err := s.Client.Ping(ctx, goodPing)
	require.NoError(s.T(), err, "must not be an error on a successful call")

	tags := tagsFromJson(s.T(), resp.Value)
	assert.Equal(s.T(), "something", tags["grpc.request.value"], "the tags should contain the correct request value")
	assert.Contains(s.T(), tags, "peer.address", "the tags should contain a peer address")
	require.Len(s.T(), tags, 2)
}

func (s *TaggingSuite) TestPing_WithNoDeadline() {
	ctx := context.TODO()
	resp, err := s.Client.Ping(ctx, goodPing)
	require.NoError(s.T(), err, "must not be an error on a successful call")

	tags := tagsFromJson(s.T(), resp.Value)
	assert.Equal(s.T(), "something", tags["grpc.request.value"], "the tags should contain the correct request value")
	assert.Contains(s.T(), tags, "peer.address", "the tags should contain a peer address")
	require.Len(s.T(), tags, 2)
}

func (s *TaggingSuite) TestPingList_WithCustomTags() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")

		tags := tagsFromJson(s.T(), resp.Value)
		assert.Equal(s.T(), "something", tags["grpc.request.value"], "the tags should contain the correct request value")
		assert.Contains(s.T(), tags, "peer.address", "the tags should contain a peer address")
	}
}

func TestTaggingOnInitialRequestSuite(t *testing.T) {
	opts := []ctxtags.Option{
		ctxtags.WithFieldExtractor(ctxtags.CodeGenRequestFieldExtractor),
	}
	// Embeds TaggingSuite as the behaviour should be identical in
	// the case of unary and server-streamed calls
	s := &ClientStreamedTaggingSuite{
		TaggingSuite: &TaggingSuite{
			InterceptorTestSuite: &grpctesting.InterceptorTestSuite{
				TestService: &tagPingBack{&grpctesting.TestPingService{T: t}},
				ServerOpts: []grpc.ServerOption{
					grpc.StreamInterceptor(ctxtags.StreamServerInterceptor(opts...)),
					grpc.UnaryInterceptor(ctxtags.UnaryServerInterceptor(opts...)),
				},
			},
		},
	}
	suite.Run(t, s)
}

type ClientStreamedTaggingSuite struct {
	*TaggingSuite
}

func (s *ClientStreamedTaggingSuite) TestPingStream_WithCustomTagsFirstRequest() {
	stream, err := s.Client.PingStream(s.SimpleCtx())
	require.NoError(s.T(), err, "should not fail on establishing the stream")

	count := 0
	for {
		switch {
		case count == 0:
			err = stream.Send(goodPing)
		case count < 3:
			err = stream.Send(anotherPing)
		default:
			err = stream.CloseSend()
		}

		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(s.T(), err, "reading stream should not fail")

		tags := tagsFromJson(s.T(), resp.Value)
		assert.Equal(s.T(), "something", tags["grpc.request.value"], "the tags should contain the correct request value")
		assert.Contains(s.T(), tags, "peer.address", "the tags should contain a peer address")
		require.Len(s.T(), tags, 2)
		count++
	}

	assert.Equal(s.T(), count, 3)
}
