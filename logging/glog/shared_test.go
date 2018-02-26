// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_glog_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type loggingPingService struct {
	pb_testproto.TestServiceServer
}

func customCodeToLevel(c codes.Code) ctx_glog.Severity {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return ctx_glog.ErrorLevel
	}
	level := grpc_glog.DefaultCodeToLevel(c)
	return level
}

func (s *loggingPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctx_glog.AddFields(ctx, ctx_glog.Fields{"custom_field": "custom_value"})
	ctx_glog.Extract(ctx).Info("some ping")
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *loggingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *loggingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	grpc_ctxtags.Extract(stream.Context()).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctx_glog.AddFields(stream.Context(), ctx_glog.Fields{"custom_field": "custom_value"})
	ctx_glog.Extract(stream.Context()).Info("some pinglist")
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *loggingPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

type glogBaseSuite struct {
	*grpc_testing.InterceptorTestSuite
	mutexBuffer *grpc_testing.MutexReadWriter
	buffer      *bytes.Buffer
	logger      grpclog.LoggerV2
}

func newGLogBaseSuite(t *testing.T) *glogBaseSuite {
	b := &bytes.Buffer{}
	muB := grpc_testing.NewMutexReadWriter(b)
	logger := grpclog.NewLoggerV2WithVerbosity(muB, ioutil.Discard, ioutil.Discard, int(ctx_glog.DebugLevel))
	return &glogBaseSuite{
		logger:      logger,
		buffer:      b,
		mutexBuffer: muB,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &loggingPingService{&grpc_testing.TestPingService{T: t}},
		},
	}
}

func (s *glogBaseSuite) SetupTest() {
	s.mutexBuffer.Lock()
	s.buffer.Reset()
	s.mutexBuffer.Unlock()
}

func (s *glogBaseSuite) getOutputJSONs() []map[string]interface{} {
	ret := make([]map[string]interface{}, 0)
	data, err := ioutil.ReadAll(s.mutexBuffer)
	if err != nil {
		return ret
	}

	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		var val map[string]interface{}
		idx := bytes.IndexRune(line, '{')
		if idx < 0 {
			continue
		}
		err := json.Unmarshal(line[idx:], &val)
		if err != nil {
			s.T().Fatalf("failed decoding output from glog JSON: %v", err)
		}

		ret = append(ret, val)
	}

	return ret
}
