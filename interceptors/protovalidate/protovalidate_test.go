// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate"
	testvalidatev1 "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func customErrorConverter(err error) error {
	return fmt.Errorf("my custom wrapper: %w", err)
}

func TestUnaryServerInterceptor(t *testing.T) {
	validator, err := protovalidate.New()
	assert.NoError(t, err)

	interceptor := protovalidate_middleware.UnaryServerInterceptor(validator)

	handler := func(ctx context.Context, req any) (any, error) {
		return "good", nil
	}

	t.Run("valid_email", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}

		resp, err := interceptor(context.TODO(), testvalidate.GoodUnaryRequest, info, handler)
		assert.NoError(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, resp, "good")
	})

	interceptor = protovalidate_middleware.UnaryServerInterceptor(validator,
		protovalidate_middleware.WithErrorConverter(customErrorConverter),
	)

	t.Run("custom_error_converter", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}

		_, err = interceptor(context.TODO(), testvalidate.BadUnaryRequest, info, handler)
		assert.Error(t, err)
		assert.Equal(t, codes.Unknown, status.Code(err))
		assert.EqualError(t, err, "my custom wrapper: validation error:\n - message: value must be a valid email address [string.email]")
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

func startGrpcServer(t *testing.T, opts ...protovalidate_middleware.Option) *grpc.ClientConn {
	lis := bufconn.Listen(bufSize)

	validator, err := protovalidate.New()
	assert.Nil(t, err)

	s := grpc.NewServer(
		grpc.StreamInterceptor(
			protovalidate_middleware.StreamServerInterceptor(validator, opts...),
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
		assert.NoError(t, err)
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
			startGrpcServer(
				t,
				protovalidate_middleware.WithIgnoreMessages(testvalidate.BadStreamRequest.ProtoReflect().Type()),
			),
		)

		out, err := client.SendStream(context.Background(), testvalidate.BadStreamRequest)
		assert.NoError(t, err)

		_, err = out.Recv()
		assert.NoError(t, err)
	})

	t.Run("custom_error_converter", func(t *testing.T) {
		client := testvalidatev1.NewTestValidateServiceClient(
			startGrpcServer(t, protovalidate_middleware.WithErrorConverter(customErrorConverter)),
		)

		out, err := client.SendStream(context.Background(), testvalidate.BadStreamRequest)
		assert.NoError(t, err)

		_, err = out.Recv()
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unknown, st.Code())
		assert.Equal(t, "my custom wrapper: validation error:\n - message: value must be a valid email address [string.email]", st.Message())
	})
}
