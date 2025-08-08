// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package prometheus

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
)

func TestServerInterceptorSuite(t *testing.T) {
	s := NewServerMetrics(WithServerHandlingTimeHistogram())
	suite.Run(t, &ServerInterceptorTestSuite{
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: &testpb.TestPingService{},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(s.StreamServerInterceptor()),
				grpc.UnaryInterceptor(s.UnaryServerInterceptor()),
			},
		},
		serverMetrics: s,
	})
}

type ServerInterceptorTestSuite struct {
	*testpb.InterceptorTestSuite

	serverMetrics *ServerMetrics
}

func (s *ServerInterceptorTestSuite) SetupTest() {
	s.serverMetrics.serverStartedCounter.Reset()
	s.serverMetrics.serverHandledCounter.Reset()
	s.serverMetrics.serverHandledHistogram.Reset()
	s.serverMetrics.serverStreamMsgReceived.Reset()
	s.serverMetrics.serverStreamMsgSent.Reset()
	s.serverMetrics.InitializeMetrics(s.Server)
}

func (s *ServerInterceptorTestSuite) TestWithSubsystem() {
	counterOpts := []CounterOption{
		WithSubsystem("subsystem1"),
	}
	histOpts := []HistogramOption{
		WithHistogramSubsystem("subsystem1"),
	}
	serverCounterOpts := WithServerCounterOptions(counterOpts...)
	serverMetrics := NewServerMetrics(serverCounterOpts, WithServerHandlingTimeHistogram(histOpts...))

	requireSubsystemName(s.T(), "subsystem1", serverMetrics.serverStartedCounter.WithLabelValues("unary", testpb.TestServiceFullName, "dummy"))
	requireHistSubsystemName(s.T(), "subsystem1", serverMetrics.serverHandledHistogram.WithLabelValues("unary", testpb.TestServiceFullName, "dummy"))
}

func (s *ServerInterceptorTestSuite) TestRegisterPresetsStuff() {
	registry := prometheus.NewPedanticRegistry()
	s.Require().NoError(registry.Register(s.serverMetrics))

	for testID, testCase := range []struct {
		metricName     string
		existingLabels []string
	}{
		// Order of label is irrelevant.
		{"grpc_server_started_total", []string{testpb.TestServiceFullName, "PingEmpty", "unary"}},
		{"grpc_server_started_total", []string{testpb.TestServiceFullName, "PingList", "server_stream"}},
		{"grpc_server_msg_received_total", []string{testpb.TestServiceFullName, "PingList", "server_stream"}},
		{"grpc_server_msg_sent_total", []string{testpb.TestServiceFullName, "PingEmpty", "unary"}},
		{"grpc_server_handling_seconds_sum", []string{testpb.TestServiceFullName, "PingEmpty", "unary"}},
		{"grpc_server_handling_seconds_count", []string{testpb.TestServiceFullName, "PingList", "server_stream"}},
		{"grpc_server_handled_total", []string{testpb.TestServiceFullName, "PingList", "server_stream", "OutOfRange"}},
		{"grpc_server_handled_total", []string{testpb.TestServiceFullName, "PingList", "server_stream", "Aborted"}},
		{"grpc_server_handled_total", []string{testpb.TestServiceFullName, "PingEmpty", "unary", "FailedPrecondition"}},
		{"grpc_server_handled_total", []string{testpb.TestServiceFullName, "PingEmpty", "unary", "ResourceExhausted"}},
	} {
		lineCount := len(fetchPrometheusLines(s.T(), registry, testCase.metricName, testCase.existingLabels...))
		s.Assert().NotZero(lineCount, "metrics must exist for test case %d", testID)
	}
}

func (s *ServerInterceptorTestSuite) TestUnaryIncrementsMetrics() {
	_, err := s.Client.PingEmpty(s.SimpleCtx(), &testpb.PingEmptyRequest{})
	s.Require().NoError(err)
	requireValue(s.T(), 1, s.serverMetrics.serverStartedCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingEmpty"))
	requireValue(s.T(), 1, s.serverMetrics.serverHandledCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingEmpty", "OK"))
	requireValueHistCount(s.T(), 1, s.serverMetrics.serverHandledHistogram.WithLabelValues("unary", testpb.TestServiceFullName, "PingEmpty"))

	_, err = s.Client.PingError(s.SimpleCtx(), &testpb.PingErrorRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)})
	s.Require().Error(err)
	requireValue(s.T(), 1, s.serverMetrics.serverStartedCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingError"))
	requireValue(s.T(), 1, s.serverMetrics.serverHandledCounter.WithLabelValues("unary", testpb.TestServiceFullName, "PingError", "FailedPrecondition"))
	requireValueHistCount(s.T(), 1, s.serverMetrics.serverHandledHistogram.WithLabelValues("unary", testpb.TestServiceFullName, "PingError"))
}

func (s *ServerInterceptorTestSuite) TestStartedStreamingIncrementsStarted() {
	_, err := s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{})
	s.Require().NoError(err)
	requireValueWithRetry(s.SimpleCtx(), s.T(), 1,
		s.serverMetrics.serverStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))

	_, err = s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)})
	s.Require().NoError(err, "PingList must not fail immediately")
	requireValueWithRetry(s.SimpleCtx(), s.T(), 2,
		s.serverMetrics.serverStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
}

func (s *ServerInterceptorTestSuite) TestStreamingIncrementsMetrics() {
	ss, _ := s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{})
	// Do a read, just for kicks.
	count := 0
	for {
		_, err := ss.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		s.Require().NoError(err, "reading pingList shouldn't fail")
		count++
	}
	s.Require().Equal(testpb.ListResponseCount, count, "Number of received msg on the wire must match")

	requireValueWithRetry(s.SimpleCtx(), s.T(), 1,
		s.serverMetrics.serverStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValueWithRetry(s.SimpleCtx(), s.T(), 1,
		s.serverMetrics.serverHandledCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList", "OK"))
	requireValueWithRetry(s.SimpleCtx(), s.T(), testpb.ListResponseCount,
		s.serverMetrics.serverStreamMsgSent.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValueWithRetry(s.SimpleCtx(), s.T(), 1,
		s.serverMetrics.serverStreamMsgReceived.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValueWithRetryHistCount(s.SimpleCtx(), s.T(), 1,
		s.serverMetrics.serverHandledHistogram.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))

	_, err := s.Client.PingList(s.SimpleCtx(), &testpb.PingListRequest{ErrorCodeReturned: uint32(codes.FailedPrecondition)}) // should return with code=FailedPrecondition
	s.Require().NoError(err, "PingList must not fail immediately")

	requireValueWithRetry(s.SimpleCtx(), s.T(), 2,
		s.serverMetrics.serverStartedCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
	requireValueWithRetry(s.SimpleCtx(), s.T(), 1,
		s.serverMetrics.serverHandledCounter.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList", "FailedPrecondition"))
	requireValueWithRetryHistCount(s.SimpleCtx(), s.T(), 2,
		s.serverMetrics.serverHandledHistogram.WithLabelValues("server_stream", testpb.TestServiceFullName, "PingList"))
}

func (s *ServerInterceptorTestSuite) TestContextCancelledTreatedAsStatus() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	stream, _ := s.Client.PingStream(ctx)
	err := stream.Send(&testpb.PingStreamRequest{})
	s.Require().NoError(err)
	cancel()

	requireValueWithRetry(s.SimpleCtx(), s.T(), 1,
		s.serverMetrics.serverHandledCounter.WithLabelValues("bidi_stream", testpb.TestServiceFullName, "PingStream", "Canceled"))
}

// fetchPrometheusLines does mocked HTTP GET request against real prometheus handler to get the same view that Prometheus
// would have while scraping this endpoint.
// Order of matching label vales does not matter.
func fetchPrometheusLines(t *testing.T, reg prometheus.Gatherer, metricName string, matchingLabelValues ...string) []string {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err, "failed creating request for Prometheus handler")

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(resp, req)
	reader := bufio.NewReader(resp.Body)

	var ret []string
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err, "error reading stuff")
		}
		if !strings.HasPrefix(line, metricName) {
			continue
		}
		matches := true
		for _, labelValue := range matchingLabelValues {
			if !strings.Contains(line, `"`+labelValue+`"`) {
				matches = false
			}
		}
		if matches {
			ret = append(ret, line)
		}

	}
	return ret
}

// toFloat64HistCount does the same thing as prometheus go client testutil.ToFloat64, but for histograms.
// TODO(bwplotka): Upstream this function to prometheus client.
func toFloat64HistCount(h prometheus.Observer) uint64 {
	var (
		m      prometheus.Metric
		mCount int
		mChan  = make(chan prometheus.Metric)
		done   = make(chan struct{})
	)

	go func() {
		for m = range mChan {
			mCount++
		}
		close(done)
	}()

	c, ok := h.(prometheus.Collector)
	if !ok {
		panic(fmt.Errorf("observer is not a collector; got: %T", h))
	}

	c.Collect(mChan)
	close(mChan)
	<-done

	if mCount != 1 {
		panic(fmt.Errorf("collected %d metrics instead of exactly 1", mCount))
	}

	pb := &dto.Metric{}
	if err := m.Write(pb); err != nil {
		panic(fmt.Errorf("metric write failed: %w", err))
	}

	if pb.Histogram != nil {
		return pb.Histogram.GetSampleCount()
	}
	panic(fmt.Errorf("collected a non-histogram metric: %s", pb))
}

func requireSubsystemName(t *testing.T, expect string, c prometheus.Collector) {
	t.Helper()
	metricFullName := reflect.ValueOf(*c.(prometheus.Metric).Desc()).FieldByName("fqName").String()

	if strings.Split(metricFullName, "_")[0] == expect {
		return
	}

	t.Errorf("expected %s value to start with %s; ", metricFullName, expect)
	t.Fail()
}

func requireNamespaceName(t *testing.T, expect string, c prometheus.Collector) {
	t.Helper()
	metricFullName := reflect.ValueOf(*c.(prometheus.Metric).Desc()).FieldByName("fqName").String()

	if strings.Split(metricFullName, "_")[0] == expect {
		return
	}

	t.Errorf("expected %s value to start with %s; ", metricFullName, expect)
	t.Fail()
}

func requireValue(t *testing.T, expect int, c prometheus.Collector) {
	t.Helper()
	v := int(testutil.ToFloat64(c))
	if v == expect {
		return
	}

	metricFullName := reflect.ValueOf(*c.(prometheus.Metric).Desc()).FieldByName("fqName").String()
	t.Errorf("expected %d %s value; got %d; ", expect, metricFullName, v)
	t.Fail()
}

func requireHistSubsystemName(t *testing.T, expect string, o prometheus.Observer) {
	t.Helper()
	metricFullName := reflect.ValueOf(*o.(prometheus.Metric).Desc()).FieldByName("fqName").String()

	if strings.Split(metricFullName, "_")[0] == expect {
		return
	}

	t.Errorf("expected %s value to start with %s; ", metricFullName, expect)
	t.Fail()
}

func requireHistNamespaceName(t *testing.T, expect string, o prometheus.Observer) {
	t.Helper()
	metricFullName := reflect.ValueOf(*o.(prometheus.Metric).Desc()).FieldByName("fqName").String()

	if strings.Split(metricFullName, "_")[0] == expect {
		return
	}

	t.Errorf("expected %s value to start with %s; ", metricFullName, expect)
	t.Fail()
}

func requireValueHistCount(t *testing.T, expect int, o prometheus.Observer) {
	t.Helper()
	v := int(toFloat64HistCount(o))
	if v == expect {
		return
	}

	metricFullName := reflect.ValueOf(*o.(prometheus.Metric).Desc()).FieldByName("fqName").String()
	t.Errorf("expected %d %s value; got %d; ", expect, metricFullName, v)
	t.Fail()
}

func requireValueWithRetry(ctx context.Context, t *testing.T, expect int, c prometheus.Collector) {
	t.Helper()
	for {
		v := int(testutil.ToFloat64(c))
		if v == expect {
			return
		}

		select {
		case <-ctx.Done():
			metricFullName := reflect.ValueOf(*c.(prometheus.Metric).Desc()).FieldByName("fqName").String()
			t.Errorf("timeout while expecting %d %s value; got %d; ", expect, metricFullName, v)
			t.Fail()
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func requireValueWithRetryHistCount(ctx context.Context, t *testing.T, expect int, o prometheus.Observer) {
	t.Helper()
	for {
		v := int(toFloat64HistCount(o))
		if v == expect {
			return
		}

		select {
		case <-ctx.Done():
			metricFullName := reflect.ValueOf(*o.(prometheus.Metric).Desc()).FieldByName("fqName").String()
			t.Errorf("timeout while expecting %d %s histogram count value; got %d; ", expect, metricFullName, v)
			t.Fail()
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
}
