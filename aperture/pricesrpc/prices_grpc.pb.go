// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pricesrpc

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

// PricesClient is the client API for Prices service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PricesClient interface {
	GetPaymentDetails(ctx context.Context, in *GetPaymentDetailsRequest, opts ...grpc.CallOption) (*GetPaymentDetailsResponse, error)
}

type pricesClient struct {
	cc grpc.ClientConnInterface
}

func NewPricesClient(cc grpc.ClientConnInterface) PricesClient {
	return &pricesClient{cc}
}

func (c *pricesClient) GetPaymentDetails(ctx context.Context, in *GetPaymentDetailsRequest, opts ...grpc.CallOption) (*GetPaymentDetailsResponse, error) {
	out := new(GetPaymentDetailsResponse)
	err := c.cc.Invoke(ctx, "/pricesrpc.Prices/GetPaymentDetails", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PricesServer is the server API for Prices service.
// All implementations must embed UnimplementedPricesServer
// for forward compatibility
type PricesServer interface {
	GetPaymentDetails(context.Context, *GetPaymentDetailsRequest) (*GetPaymentDetailsResponse, error)
	mustEmbedUnimplementedPricesServer()
}

// UnimplementedPricesServer must be embedded to have forward compatible implementations.
type UnimplementedPricesServer struct {
}

func (UnimplementedPricesServer) GetPaymentDetails(context.Context, *GetPaymentDetailsRequest) (*GetPaymentDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPaymentDetails not implemented")
}
func (UnimplementedPricesServer) mustEmbedUnimplementedPricesServer() {}

// UnsafePricesServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PricesServer will
// result in compilation errors.
type UnsafePricesServer interface {
	mustEmbedUnimplementedPricesServer()
}

func RegisterPricesServer(s grpc.ServiceRegistrar, srv PricesServer) {
	s.RegisterService(&Prices_ServiceDesc, srv)
}

func _Prices_GetPaymentDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPaymentDetailsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PricesServer).GetPaymentDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pricesrpc.Prices/GetPaymentDetails",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PricesServer).GetPaymentDetails(ctx, req.(*GetPaymentDetailsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Prices_ServiceDesc is the grpc.ServiceDesc for Prices service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Prices_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pricesrpc.Prices",
	HandlerType: (*PricesServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPaymentDetails",
			Handler:    _Prices_GetPaymentDetails_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "prices.proto",
}
