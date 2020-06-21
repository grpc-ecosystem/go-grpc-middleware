package grpctesting

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testpb"
)

func TestPingServiceOnWire(t *testing.T) {
	stopped := make(chan error)
	serverListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "must be able to allocate a port for serverListener")

	server := grpc.NewServer()
	testpb.RegisterTestServiceServer(server, &TestPingService{T: t})

	go func() {
		defer close(stopped)
		stopped <- server.Serve(serverListener)
	}()
	defer func() {
		server.Stop()
		<-stopped
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This is the point where we hook up the interceptor.
	clientConn, err := grpc.DialContext(
		ctx,
		serverListener.Addr().String(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	require.NoError(t, err, "must not error on client Dial")

	testClient := testpb.NewTestServiceClient(clientConn)
	select {
	case err := <-stopped:
		t.Fatal("gRPC server stopped prematurely", err)
	default:
	}

	r, err := testClient.PingEmpty(context.Background(), &testpb.Empty{})
	require.NoError(t, err)
	require.Equal(t, "default_response_value", r.Value)
	require.Equal(t, int32(0), r.Counter)

	r2, err := testClient.Ping(context.Background(), &testpb.PingRequest{Value: "24"})
	require.NoError(t, err)
	require.Equal(t, "24", r2.Value)
	require.Equal(t, int32(0), r2.Counter)

	_, err = testClient.PingError(context.Background(), &testpb.PingRequest{Value: "24"})
	require.Error(t, err)

	l, err := testClient.PingList(context.Background(), &testpb.PingRequest{Value: "24"})
	require.NoError(t, err)
	for i := 0; i < ListResponseCount; i++ {
		r, err := l.Recv()
		require.NoError(t, err)
		require.Equal(t, "24", r.Value)
		require.Equal(t, int32(i), r.Counter)
	}

	s, err := testClient.PingStream(context.Background())
	require.NoError(t, err)
	for i := 0; i < ListResponseCount; i++ {
		require.NoError(t, s.Send(&testpb.PingRequest{Value: fmt.Sprintf("%v", i)}))

		r, err := s.Recv()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%v", i), r.Value)
		require.Equal(t, int32(i), r.Counter)
	}

	select {
	case err := <-stopped:
		t.Fatal("gRPC server stopped prematurely", err)
	default:
	}
}
