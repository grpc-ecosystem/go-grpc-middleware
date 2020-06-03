package grpc_zerolog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	grpc_zerolog "github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzerolog"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type loggingPingService struct {
	pb_testproto.TestServiceServer
}

func customCodeToLevel(c codes.Code) zerolog.Level {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return zerolog.ErrorLevel
	}
	level := grpc_zerolog.DefaultCodeToLevel(c)
	return level
}

func (s *loggingPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctxzerolog.AddFields(ctx, map[string]interface{}{"custom_field": "custom_value"})
	l := ctxzerolog.Extract(ctx).Logger()
	l.Info().Msg("some ping")
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *loggingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *loggingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	grpc_ctxtags.Extract(stream.Context()).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctxzerolog.AddFields(stream.Context(), map[string]interface{}{"custom_field": "custom_value"})
	l := ctxzerolog.Extract(stream.Context()).Logger()
	l.Info().Msg("some pinglist")
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *loggingPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

type zerologBaseSuite struct {
	*grpc_testing.InterceptorTestSuite
	mutexBuffer *grpc_testing.MutexReadWriter
	buffer      *bytes.Buffer
	logger      *zerolog.Logger
}

func newZerologBaseSuite(t *testing.T) *zerologBaseSuite {
	b := &bytes.Buffer{}
	muB := grpc_testing.NewMutexReadWriter(b)
	logger := zerolog.New(muB)
	return &zerologBaseSuite{
		logger:      &logger,
		buffer:      b,
		mutexBuffer: muB,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &loggingPingService{&grpc_testing.TestPingService{T: t}},
		},
	}
}

func (s *zerologBaseSuite) SetupTest() {
	s.mutexBuffer.Lock()
	s.buffer.Reset()
	s.mutexBuffer.Unlock()
}

func (s *zerologBaseSuite) getOutputJSONs() []map[string]interface{} {
	ret := make([]map[string]interface{}, 0)
	dec := json.NewDecoder(s.mutexBuffer)

	for {
		var val map[string]interface{}
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.T().Fatalf("failed decoding output from Zerolog JSON: %v", err)
		}

		ret = append(ret, val)
	}

	return ret
}

func StubMessageProducer(ctx context.Context, _ string, level zerolog.Level, code codes.Code, err error, fields map[string]interface{}) {
	format := "custom message"
	ctxLogger := ctxzerolog.Extract(ctx).Logger()
	event := ctxLogger.WithLevel(level)
	if err != nil {
		event = event.Err(err)
	}
	event.Fields(fields).Msg(format)
}
