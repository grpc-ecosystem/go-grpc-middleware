// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate_test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/reflect/protoreflect"

	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate"
	testvalidatev1 "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate/v1"
)

func TestUnaryServerInterceptor(t *testing.T) {
	validator, err := protovalidate.New()
	assert.Nil(t, err)

	interceptor := protovalidate_middleware.UnaryServerInterceptor(validator)

	handler := func(ctx context.Context, req any) (any, error) {
		return "good", nil
	}

	t.Run("valid_email", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}

		resp, err := interceptor(context.TODO(), testvalidate.GoodUnaryRequest, info, handler)
		assert.Nil(t, err)
		assert.Equal(t, resp, "good")
	})

	t.Run("invalid_email", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}

		_, err = interceptor(context.TODO(), testvalidate.BadUnaryRequest, info, handler)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	interceptor = protovalidate_middleware.UnaryServerInterceptor(validator,
		protovalidate_middleware.WithIgnoreMessages(testvalidate.BadUnaryRequest.ProtoReflect().Type()),
	)

	t.Run("invalid_email_ignored", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}

		resp, err := interceptor(context.TODO(), testvalidate.BadUnaryRequest, info, handler)
		assert.Nil(t, err)
		assert.Equal(t, resp, "good")
	})
}

type server struct {
	testvalidatev1.UnimplementedTestValidateServiceServer
}

func (g *server) SendStream(
	_ *testvalidatev1.SendStreamRequest,
	stream testvalidatev1.TestValidateService_SendStreamServer,
) error {
	if err := stream.Send(&testvalidatev1.SendStreamResponse{}); err != nil {
		return err
	}

	return nil
}

const bufSize = 1024 * 1024

func startGrpcServer(t *testing.T, ignoreMessages ...protoreflect.MessageType) *grpc.ClientConn {
	lis := bufconn.Listen(bufSize)

	validator, err := protovalidate.New()
	assert.Nil(t, err)

	s := grpc.NewServer(
		grpc.StreamInterceptor(
			protovalidate_middleware.StreamServerInterceptor(validator,
				protovalidate_middleware.WithIgnoreMessages(ignoreMessages...),
			),
		),
	)
	testvalidatev1.RegisterTestValidateServiceServer(s, &server{})
	go func() {
		if err = s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		s.GracefulStop()
		_ = lis.Close()
	})

	return conn
}

func TestStreamServerInterceptor(t *testing.T) {
	t.Run("valid_email", func(t *testing.T) {
		client := testvalidatev1.NewTestValidateServiceClient(
			startGrpcServer(t),
		)

		_, err := client.SendStream(context.Background(), testvalidate.GoodStreamRequest)
		assert.Nil(t, err)
	})

	t.Run("invalid_email", func(t *testing.T) {
		client := testvalidatev1.NewTestValidateServiceClient(
			startGrpcServer(t),
		)

		out, err := client.SendStream(context.Background(), testvalidate.BadStreamRequest)
		assert.Nil(t, err)

		_, err = out.Recv()
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid_email_ignored", func(t *testing.T) {
		client := testvalidatev1.NewTestValidateServiceClient(
			startGrpcServer(t, testvalidate.BadStreamRequest.ProtoReflect().Type()),
		)

		out, err := client.SendStream(context.Background(), testvalidate.BadStreamRequest)
		assert.Nil(t, err)

		_, err = out.Recv()
		assert.Nil(t, err)
	})
}
