// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package blocks

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

// BlocksServiceClient is the client API for BlocksService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlocksServiceClient interface {
	New(ctx context.Context, in *Block, opts ...grpc.CallOption) (*Block, error)
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Block, error)
	ManualAction(ctx context.Context, in *ManualRequest, opts ...grpc.CallOption) (*ManualResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
}

type blocksServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBlocksServiceClient(cc grpc.ClientConnInterface) BlocksServiceClient {
	return &blocksServiceClient{cc}
}

func (c *blocksServiceClient) New(ctx context.Context, in *Block, opts ...grpc.CallOption) (*Block, error) {
	out := new(Block)
	err := c.cc.Invoke(ctx, "/trader.blocks.BlocksService/New", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/trader.blocks.BlocksService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Block, error) {
	out := new(Block)
	err := c.cc.Invoke(ctx, "/trader.blocks.BlocksService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) ManualAction(ctx context.Context, in *ManualRequest, opts ...grpc.CallOption) (*ManualResponse, error) {
	out := new(ManualResponse)
	err := c.cc.Invoke(ctx, "/trader.blocks.BlocksService/ManualAction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/trader.blocks.BlocksService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlocksServiceServer is the server API for BlocksService service.
// All implementations must embed UnimplementedBlocksServiceServer
// for forward compatibility
type BlocksServiceServer interface {
	New(context.Context, *Block) (*Block, error)
	List(context.Context, *ListRequest) (*ListResponse, error)
	Get(context.Context, *GetRequest) (*Block, error)
	ManualAction(context.Context, *ManualRequest) (*ManualResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	mustEmbedUnimplementedBlocksServiceServer()
}

// UnimplementedBlocksServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBlocksServiceServer struct {
}

func (UnimplementedBlocksServiceServer) New(context.Context, *Block) (*Block, error) {
	return nil, status.Errorf(codes.Unimplemented, "method New not implemented")
}
func (UnimplementedBlocksServiceServer) List(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedBlocksServiceServer) Get(context.Context, *GetRequest) (*Block, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedBlocksServiceServer) ManualAction(context.Context, *ManualRequest) (*ManualResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ManualAction not implemented")
}
func (UnimplementedBlocksServiceServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedBlocksServiceServer) mustEmbedUnimplementedBlocksServiceServer() {}

// UnsafeBlocksServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlocksServiceServer will
// result in compilation errors.
type UnsafeBlocksServiceServer interface {
	mustEmbedUnimplementedBlocksServiceServer()
}

func RegisterBlocksServiceServer(s grpc.ServiceRegistrar, srv BlocksServiceServer) {
	s.RegisterService(&BlocksService_ServiceDesc, srv)
}

func _BlocksService_New_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Block)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).New(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trader.blocks.BlocksService/New",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).New(ctx, req.(*Block))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trader.blocks.BlocksService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trader.blocks.BlocksService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_ManualAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ManualRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).ManualAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trader.blocks.BlocksService/ManualAction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).ManualAction(ctx, req.(*ManualRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trader.blocks.BlocksService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BlocksService_ServiceDesc is the grpc.ServiceDesc for BlocksService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BlocksService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "trader.blocks.BlocksService",
	HandlerType: (*BlocksServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "New",
			Handler:    _BlocksService_New_Handler,
		},
		{
			MethodName: "List",
			Handler:    _BlocksService_List_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _BlocksService_Get_Handler,
		},
		{
			MethodName: "ManualAction",
			Handler:    _BlocksService_ManualAction_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _BlocksService_Delete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "blocks.proto",
}
