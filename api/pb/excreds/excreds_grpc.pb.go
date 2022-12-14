// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package excreds

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

// ExCredsServiceClient is the client API for ExCredsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExCredsServiceClient interface {
	New(ctx context.Context, in *ExchangeCreds, opts ...grpc.CallOption) (*ExchangeCreds, error)
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*ExchangeCreds, error)
}

type exCredsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewExCredsServiceClient(cc grpc.ClientConnInterface) ExCredsServiceClient {
	return &exCredsServiceClient{cc}
}

func (c *exCredsServiceClient) New(ctx context.Context, in *ExchangeCreds, opts ...grpc.CallOption) (*ExchangeCreds, error) {
	out := new(ExchangeCreds)
	err := c.cc.Invoke(ctx, "/ataas.excreds.ExCredsService/New", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exCredsServiceClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/ataas.excreds.ExCredsService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exCredsServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/ataas.excreds.ExCredsService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exCredsServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*ExchangeCreds, error) {
	out := new(ExchangeCreds)
	err := c.cc.Invoke(ctx, "/ataas.excreds.ExCredsService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExCredsServiceServer is the server API for ExCredsService service.
// All implementations must embed UnimplementedExCredsServiceServer
// for forward compatibility
type ExCredsServiceServer interface {
	New(context.Context, *ExchangeCreds) (*ExchangeCreds, error)
	List(context.Context, *ListRequest) (*ListResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	Get(context.Context, *GetRequest) (*ExchangeCreds, error)
	mustEmbedUnimplementedExCredsServiceServer()
}

// UnimplementedExCredsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedExCredsServiceServer struct {
}

func (UnimplementedExCredsServiceServer) New(context.Context, *ExchangeCreds) (*ExchangeCreds, error) {
	return nil, status.Errorf(codes.Unimplemented, "method New not implemented")
}
func (UnimplementedExCredsServiceServer) List(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedExCredsServiceServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedExCredsServiceServer) Get(context.Context, *GetRequest) (*ExchangeCreds, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedExCredsServiceServer) mustEmbedUnimplementedExCredsServiceServer() {}

// UnsafeExCredsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExCredsServiceServer will
// result in compilation errors.
type UnsafeExCredsServiceServer interface {
	mustEmbedUnimplementedExCredsServiceServer()
}

func RegisterExCredsServiceServer(s grpc.ServiceRegistrar, srv ExCredsServiceServer) {
	s.RegisterService(&ExCredsService_ServiceDesc, srv)
}

func _ExCredsService_New_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExchangeCreds)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExCredsServiceServer).New(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ataas.excreds.ExCredsService/New",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExCredsServiceServer).New(ctx, req.(*ExchangeCreds))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExCredsService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExCredsServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ataas.excreds.ExCredsService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExCredsServiceServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExCredsService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExCredsServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ataas.excreds.ExCredsService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExCredsServiceServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExCredsService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExCredsServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ataas.excreds.ExCredsService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExCredsServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ExCredsService_ServiceDesc is the grpc.ServiceDesc for ExCredsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ExCredsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ataas.excreds.ExCredsService",
	HandlerType: (*ExCredsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "New",
			Handler:    _ExCredsService_New_Handler,
		},
		{
			MethodName: "List",
			Handler:    _ExCredsService_List_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _ExCredsService_Delete_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _ExCredsService_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "excreds.proto",
}
