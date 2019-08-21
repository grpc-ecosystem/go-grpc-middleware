package grpc_zerolog_test

import (
	"bytes"
	"encoding/json"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzr"
	"io"
	"testing"

	"context"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/rs/zerolog"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type loggingPingService struct {
	pb_testproto.TestServiceServer
}

func (s *loggingPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctxzr.AddFields(ctx, []interface{}{"custom_field", "custom_value"})
	var ctxLog = ctxzr.Extract(ctx)
	ctxLog.Fields = append(ctxLog.Fields, "msg", "some ping")
	return s.TestServiceServer.Ping(ctxzr.ToContext(ctx, ctxLog), ping)
}

func (s *loggingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *loggingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	grpc_ctxtags.Extract(stream.Context()).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctxzr.AddFields(stream.Context(), []interface{}{"custom_field", "custom_value"})
	var ctxLog = ctxzr.Extract(stream.Context())
	ctxLog.Fields = append(ctxLog.Fields, "msg", "some pinglist")
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *loggingPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

type ZRBaseSuite struct {
	*grpc_testing.InterceptorTestSuite
	mutexBuffer *grpc_testing.MutexReadWriter
	buffer      *bytes.Buffer
	logger      *ctxzr.CtxLogger
}

func newZRBaseSuite(t *testing.T) *ZRBaseSuite {
	b := &bytes.Buffer{}
	muB := grpc_testing.NewMutexReadWriter(b)
	logger := zerolog.New(muB)
	fields := []interface{}{}
	return &ZRBaseSuite{
		logger:      &ctxzr.CtxLogger{Logger: &logger, Fields: fields},
		buffer:      b,
		mutexBuffer: muB,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &loggingPingService{&grpc_testing.TestPingService{T: t}},
		},
	}
}

func (s *ZRBaseSuite) SetupTest() {
	s.mutexBuffer.Lock()
	s.buffer.Reset()
	s.mutexBuffer.Unlock()
}

func (s *ZRBaseSuite) getOutputJSONs() []map[string]interface{} {
	ret := make([]map[string]interface{}, 0)
	dec := json.NewDecoder(s.mutexBuffer)

	for {
		var val map[string]interface{}
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.T().Fatalf("failed decoding output from go-zerolog JSON: %v", err)
		}

		ret = append(ret, val)
	}

	return ret
}
