// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: testing/testvalidate/v1/test_validate.proto

package testvalidatev1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	TestValidateService_Send_FullMethodName       = "/testing.testvalidate.v1.TestValidateService/Send"
	TestValidateService_SendStream_FullMethodName = "/testing.testvalidate.v1.TestValidateService/SendStream"
)

// TestValidateServiceClient is the client API for TestValidateService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TestValidateServiceClient interface {
	Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error)
	SendStream(ctx context.Context, in *SendStreamRequest, opts ...grpc.CallOption) (TestValidateService_SendStreamClient, error)
}

type testValidateServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTestValidateServiceClient(cc grpc.ClientConnInterface) TestValidateServiceClient {
	return &testValidateServiceClient{cc}
}

func (c *testValidateServiceClient) Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error) {
	out := new(SendResponse)
	err := c.cc.Invoke(ctx, TestValidateService_Send_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testValidateServiceClient) SendStream(ctx context.Context, in *SendStreamRequest, opts ...grpc.CallOption) (TestValidateService_SendStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &TestValidateService_ServiceDesc.Streams[0], TestValidateService_SendStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &testValidateServiceSendStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type TestValidateService_SendStreamClient interface {
	Recv() (*SendStreamResponse, error)
	grpc.ClientStream
}

type testValidateServiceSendStreamClient struct {
	grpc.ClientStream
}

func (x *testValidateServiceSendStreamClient) Recv() (*SendStreamResponse, error) {
	m := new(SendStreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TestValidateServiceServer is the server API for TestValidateService service.
// All implementations should embed UnimplementedTestValidateServiceServer
// for forward compatibility
type TestValidateServiceServer interface {
	Send(context.Context, *SendRequest) (*SendResponse, error)
	SendStream(*SendStreamRequest, TestValidateService_SendStreamServer) error
}

// UnimplementedTestValidateServiceServer should be embedded to have forward compatible implementations.
type UnimplementedTestValidateServiceServer struct {
}

func (UnimplementedTestValidateServiceServer) Send(context.Context, *SendRequest) (*SendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedTestValidateServiceServer) SendStream(*SendStreamRequest, TestValidateService_SendStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method SendStream not implemented")
}

// UnsafeTestValidateServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestValidateServiceServer will
// result in compilation errors.
type UnsafeTestValidateServiceServer interface {
	mustEmbedUnimplementedTestValidateServiceServer()
}

func RegisterTestValidateServiceServer(s grpc.ServiceRegistrar, srv TestValidateServiceServer) {
	s.RegisterService(&TestValidateService_ServiceDesc, srv)
}

func _TestValidateService_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestValidateServiceServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TestValidateService_Send_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestValidateServiceServer).Send(ctx, req.(*SendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TestValidateService_SendStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SendStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TestValidateServiceServer).SendStream(m, &testValidateServiceSendStreamServer{stream})
}

type TestValidateService_SendStreamServer interface {
	Send(*SendStreamResponse) error
	grpc.ServerStream
}

type testValidateServiceSendStreamServer struct {
	grpc.ServerStream
}

func (x *testValidateServiceSendStreamServer) Send(m *SendStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

// TestValidateService_ServiceDesc is the grpc.ServiceDesc for TestValidateService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TestValidateService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "testing.testvalidate.v1.TestValidateService",
	HandlerType: (*TestValidateServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _TestValidateService_Send_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendStream",
			Handler:       _TestValidateService_SendStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "testing/testvalidate/v1/test_validate.proto",
}
