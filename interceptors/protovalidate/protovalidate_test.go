// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package protovalidate_test

import (
	"context"
	"log"
	"net"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protovalidate "github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate"
	testvalidatev1 "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testvalidate/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestUnaryServerInterceptor(t *testing.T) {
	validator, err := protovalidate.New()
	assert.Nil(t, err)

	interceptor := protovalidate_middleware.UnaryServerInterceptor(validator)

	handler := func(ctx context.Context, req any) (any, error) {
		return "good", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "FakeMethod"}

	t.Run("valid_email", func(t *testing.T) {
		resp, err := interceptor(context.TODO(), testvalidate.GoodUnaryRequest, info, handler)
		assert.Nil(t, err)
		assert.Equal(t, resp, "good")
	})

	t.Run("invalid_email", func(t *testing.T) {
		_, err = interceptor(context.TODO(), testvalidate.BadUnaryRequest, info, handler)
		assertEqualViolation(t, &validate.Violation{
			Field: &validate.FieldPath{
				Elements: []*validate.FieldPathElement{
					{
						FieldNumber: proto.Int32(1),
						FieldName:   proto.String("message"),
						FieldType:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
				},
			},
			Rule: &validate.FieldPath{
				Elements: []*validate.FieldPathElement{
					{
						FieldNumber: proto.Int32(14),
						FieldName:   proto.String("string"),
						FieldType:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					},
					{
						FieldNumber: proto.Int32(12),
						FieldName:   proto.String("email"),
						FieldType:   descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(),
					},
				},
			},
			ConstraintId: proto.String("string.email"),
			Message:      proto.String("value must be a valid email address"),
		}, err)
	})

	t.Run("not_protobuf", func(t *testing.T) {
		_, err = interceptor(context.Background(), "not protobuf", info, handler)
		assert.EqualError(t, err, "rpc error: code = Internal desc = unsupported message type: string")
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	interceptor = protovalidate_middleware.UnaryServerInterceptor(validator,
		protovalidate_middleware.WithIgnoreMessages(testvalidate.BadUnaryRequest.ProtoReflect().Type()),
	)

	t.Run("invalid_email_ignored", func(t *testing.T) {
		resp, err := interceptor(context.TODO(), testvalidate.BadUnaryRequest, info, handler)
		assert.Nil(t, err)
		assert.Equal(t, resp, "good")
	})
}

type server struct {
	testvalidatev1.UnimplementedTestValidateServiceServer

	called *bool
}

func (g *server) SendStream(
	_ *testvalidatev1.SendStreamRequest,
	stream testvalidatev1.TestValidateService_SendStreamServer,
) error {
	*g.called = true
	if err := stream.Send(&testvalidatev1.SendStreamResponse{}); err != nil {
		return err
	}

	return nil
}

const bufSize = 1024 * 1024

func startGrpcServer(t *testing.T, called *bool, ignoreMessages ...protoreflect.MessageType) *grpc.ClientConn {
	lis := bufconn.Listen(bufSize)

	validator, err := protovalidate.New()
	require.Nil(t, err)

	s := grpc.NewServer(
		grpc.StreamInterceptor(
			protovalidate_middleware.StreamServerInterceptor(validator,
				protovalidate_middleware.WithIgnoreMessages(ignoreMessages...),
			),
		),
	)
	testvalidatev1.RegisterTestValidateServiceServer(s, &server{called: called})
	go func() {
		if err = s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
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
		called := proto.Bool(false)
		client := testvalidatev1.NewTestValidateServiceClient(
			startGrpcServer(t, called),
		)

		out, err := client.SendStream(context.Background(), testvalidate.GoodStreamRequest)
		require.Nil(t, err, "SendStream failed: %v", err)

		_, err = out.Recv()
		t.Log(err)
		assert.Nil(t, err)
		assert.True(t, *called)
	})

	t.Run("invalid_email", func(t *testing.T) {
		called := proto.Bool(false)
		client := testvalidatev1.NewTestValidateServiceClient(
			startGrpcServer(t, called),
		)

		out, err := client.SendStream(context.Background(), testvalidate.BadStreamRequest)
		require.Nil(t, err, "SendStream failed: %v", err)

		_, err = out.Recv()
		assertEqualViolation(t, &validate.Violation{
			Field: &validate.FieldPath{
				Elements: []*validate.FieldPathElement{
					{
						FieldNumber: proto.Int32(1),
						FieldName:   proto.String("message"),
						FieldType:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
				},
			},
			Rule: &validate.FieldPath{
				Elements: []*validate.FieldPathElement{
					{
						FieldNumber: proto.Int32(14),
						FieldName:   proto.String("string"),
						FieldType:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					},
					{
						FieldNumber: proto.Int32(12),
						FieldName:   proto.String("email"),
						FieldType:   descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(),
					},
				},
			},
			ConstraintId: proto.String("string.email"),
			Message:      proto.String("value must be a valid email address"),
		}, err)
		assert.False(t, *called)
	})

	t.Run("invalid_email_ignored", func(t *testing.T) {
		called := proto.Bool(false)
		client := testvalidatev1.NewTestValidateServiceClient(
			startGrpcServer(t, called, testvalidate.BadStreamRequest.ProtoReflect().Type()),
		)

		out, err := client.SendStream(context.Background(), testvalidate.BadStreamRequest)
		require.Nil(t, err, "SendStream failed: %v", err)

		_, err = out.Recv()
		assert.Nil(t, err)
		assert.True(t, *called)
	})
}

func assertEqualViolation(tb testing.TB, want *validate.Violation, got error) bool {
	require.Error(tb, got)
	st := status.Convert(got)
	assert.Equal(tb, codes.InvalidArgument, st.Code())
	details := st.Proto().GetDetails()
	require.Len(tb, details, 1)
	gotpb, unwrapErr := details[0].UnmarshalNew()
	require.Nil(tb, unwrapErr)
	violations := &validate.Violations{
		Violations: []*validate.Violation{want},
	}
	tb.Logf("got: %v", gotpb)
	tb.Logf("want: %v", violations)
	return assert.True(tb, proto.Equal(gotpb, violations))
}
