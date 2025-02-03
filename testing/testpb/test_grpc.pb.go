// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v4.22.0
// source: test.proto

package testpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	TestService_PingEmpty_FullMethodName        = "/testing.testpb.v1.TestService/PingEmpty"
	TestService_Ping_FullMethodName             = "/testing.testpb.v1.TestService/Ping"
	TestService_PingError_FullMethodName        = "/testing.testpb.v1.TestService/PingError"
	TestService_PingList_FullMethodName         = "/testing.testpb.v1.TestService/PingList"
	TestService_PingStream_FullMethodName       = "/testing.testpb.v1.TestService/PingStream"
	TestService_PingClientStream_FullMethodName = "/testing.testpb.v1.TestService/PingClientStream"
)

// TestServiceClient is the client API for TestService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TestServiceClient interface {
	PingEmpty(ctx context.Context, in *PingEmptyRequest, opts ...grpc.CallOption) (*PingEmptyResponse, error)
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	PingError(ctx context.Context, in *PingErrorRequest, opts ...grpc.CallOption) (*PingErrorResponse, error)
	PingList(ctx context.Context, in *PingListRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[PingListResponse], error)
	PingStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[PingStreamRequest, PingStreamResponse], error)
	PingClientStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[PingClientStreamRequest, PingClientStreamResponse], error)
}

type testServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTestServiceClient(cc grpc.ClientConnInterface) TestServiceClient {
	return &testServiceClient{cc}
}

func (c *testServiceClient) PingEmpty(ctx context.Context, in *PingEmptyRequest, opts ...grpc.CallOption) (*PingEmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingEmptyResponse)
	err := c.cc.Invoke(ctx, TestService_PingEmpty_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, TestService_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testServiceClient) PingError(ctx context.Context, in *PingErrorRequest, opts ...grpc.CallOption) (*PingErrorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingErrorResponse)
	err := c.cc.Invoke(ctx, TestService_PingError_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testServiceClient) PingList(ctx context.Context, in *PingListRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[PingListResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &TestService_ServiceDesc.Streams[0], TestService_PingList_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[PingListRequest, PingListResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TestService_PingListClient = grpc.ServerStreamingClient[PingListResponse]

func (c *testServiceClient) PingStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[PingStreamRequest, PingStreamResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &TestService_ServiceDesc.Streams[1], TestService_PingStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[PingStreamRequest, PingStreamResponse]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TestService_PingStreamClient = grpc.BidiStreamingClient[PingStreamRequest, PingStreamResponse]

func (c *testServiceClient) PingClientStream(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[PingClientStreamRequest, PingClientStreamResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &TestService_ServiceDesc.Streams[2], TestService_PingClientStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[PingClientStreamRequest, PingClientStreamResponse]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TestService_PingClientStreamClient = grpc.ClientStreamingClient[PingClientStreamRequest, PingClientStreamResponse]

// TestServiceServer is the server API for TestService service.
// All implementations must embed UnimplementedTestServiceServer
// for forward compatibility.
type TestServiceServer interface {
	PingEmpty(context.Context, *PingEmptyRequest) (*PingEmptyResponse, error)
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	PingError(context.Context, *PingErrorRequest) (*PingErrorResponse, error)
	PingList(*PingListRequest, grpc.ServerStreamingServer[PingListResponse]) error
	PingStream(grpc.BidiStreamingServer[PingStreamRequest, PingStreamResponse]) error
	PingClientStream(grpc.ClientStreamingServer[PingClientStreamRequest, PingClientStreamResponse]) error
	mustEmbedUnimplementedTestServiceServer()
}

// UnimplementedTestServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTestServiceServer struct{}

func (UnimplementedTestServiceServer) PingEmpty(context.Context, *PingEmptyRequest) (*PingEmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingEmpty not implemented")
}
func (UnimplementedTestServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedTestServiceServer) PingError(context.Context, *PingErrorRequest) (*PingErrorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingError not implemented")
}
func (UnimplementedTestServiceServer) PingList(*PingListRequest, grpc.ServerStreamingServer[PingListResponse]) error {
	return status.Errorf(codes.Unimplemented, "method PingList not implemented")
}
func (UnimplementedTestServiceServer) PingStream(grpc.BidiStreamingServer[PingStreamRequest, PingStreamResponse]) error {
	return status.Errorf(codes.Unimplemented, "method PingStream not implemented")
}
func (UnimplementedTestServiceServer) PingClientStream(grpc.ClientStreamingServer[PingClientStreamRequest, PingClientStreamResponse]) error {
	return status.Errorf(codes.Unimplemented, "method PingClientStream not implemented")
}
func (UnimplementedTestServiceServer) mustEmbedUnimplementedTestServiceServer() {}
func (UnimplementedTestServiceServer) testEmbeddedByValue()                     {}

// UnsafeTestServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestServiceServer will
// result in compilation errors.
type UnsafeTestServiceServer interface {
	mustEmbedUnimplementedTestServiceServer()
}

func RegisterTestServiceServer(s grpc.ServiceRegistrar, srv TestServiceServer) {
	// If the following call pancis, it indicates UnimplementedTestServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TestService_ServiceDesc, srv)
}

func _TestService_PingEmpty_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingEmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServiceServer).PingEmpty(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TestService_PingEmpty_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServiceServer).PingEmpty(ctx, req.(*PingEmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TestService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TestService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TestService_PingError_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingErrorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServiceServer).PingError(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TestService_PingError_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServiceServer).PingError(ctx, req.(*PingErrorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TestService_PingList_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PingListRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TestServiceServer).PingList(m, &grpc.GenericServerStream[PingListRequest, PingListResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TestService_PingListServer = grpc.ServerStreamingServer[PingListResponse]

func _TestService_PingStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TestServiceServer).PingStream(&grpc.GenericServerStream[PingStreamRequest, PingStreamResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TestService_PingStreamServer = grpc.BidiStreamingServer[PingStreamRequest, PingStreamResponse]

func _TestService_PingClientStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TestServiceServer).PingClientStream(&grpc.GenericServerStream[PingClientStreamRequest, PingClientStreamResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TestService_PingClientStreamServer = grpc.ClientStreamingServer[PingClientStreamRequest, PingClientStreamResponse]

// TestService_ServiceDesc is the grpc.ServiceDesc for TestService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TestService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "testing.testpb.v1.TestService",
	HandlerType: (*TestServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PingEmpty",
			Handler:    _TestService_PingEmpty_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _TestService_Ping_Handler,
		},
		{
			MethodName: "PingError",
			Handler:    _TestService_PingError_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PingList",
			Handler:       _TestService_PingList_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "PingStream",
			Handler:       _TestService_PingStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "PingClientStream",
			Handler:       _TestService_PingClientStream_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "test.proto",
}
