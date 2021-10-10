// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package notify

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

// NotifyServiceClient is the client API for NotifyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NotifyServiceClient interface {
	Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error)
}

type notifyServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotifyServiceClient(cc grpc.ClientConnInterface) NotifyServiceClient {
	return &notifyServiceClient{cc}
}

func (c *notifyServiceClient) Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error) {
	out := new(SendResponse)
	err := c.cc.Invoke(ctx, "/ataas.notify.NotifyService/Send", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotifyServiceServer is the server API for NotifyService service.
// All implementations must embed UnimplementedNotifyServiceServer
// for forward compatibility
type NotifyServiceServer interface {
	Send(context.Context, *SendRequest) (*SendResponse, error)
	mustEmbedUnimplementedNotifyServiceServer()
}

// UnimplementedNotifyServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNotifyServiceServer struct {
}

func (UnimplementedNotifyServiceServer) Send(context.Context, *SendRequest) (*SendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedNotifyServiceServer) mustEmbedUnimplementedNotifyServiceServer() {}

// UnsafeNotifyServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NotifyServiceServer will
// result in compilation errors.
type UnsafeNotifyServiceServer interface {
	mustEmbedUnimplementedNotifyServiceServer()
}

func RegisterNotifyServiceServer(s grpc.ServiceRegistrar, srv NotifyServiceServer) {
	s.RegisterService(&NotifyService_ServiceDesc, srv)
}

func _NotifyService_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyServiceServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ataas.notify.NotifyService/Send",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyServiceServer).Send(ctx, req.(*SendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NotifyService_ServiceDesc is the grpc.ServiceDesc for NotifyService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NotifyService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ataas.notify.NotifyService",
	HandlerType: (*NotifyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _NotifyService_Send_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "notify.proto",
}
