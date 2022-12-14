package api

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"pm.tcfw.com.au/source/ataas/api/pb/blocks"
	"pm.tcfw.com.au/source/ataas/api/pb/excreds"
	"pm.tcfw.com.au/source/ataas/api/pb/orders"
	"pm.tcfw.com.au/source/ataas/api/pb/passport"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/api/pb/users"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

func newRouter(ctx context.Context) (*runtime.ServeMux, error) {
	r := runtime.NewServeMux(
		// runtime.WithMarshalerOption("application/json", &marshalers.JSONMarshaler{}),
		runtime.WithForwardResponseOption(httpResponseModifier),
		runtime.WithOutgoingHeaderMatcher(httpHeaderMatch),
	)

	conn, err := grpc.DialContext(
		ctx,
		viper.GetString("grpc.addr"),
		rpcUtils.InternalClientOptions()...,
	)
	if err != nil {
		return nil, err
	}

	if err := registerLocalBlocks(ctx, r, conn); err != nil {
		return nil, err
	}

	if err := registerLocalOrders(ctx, r, conn); err != nil {
		return nil, err
	}

	if err := registerLocalPassport(ctx, r, conn); err != nil {
		return nil, err
	}

	if err := registerLocalTicks(ctx, r, conn); err != nil {
		return nil, err
	}

	if err := registerLocalUsers(ctx, r, conn); err != nil {
		return nil, err
	}

	if err := registerLocalStrategy(ctx, r, conn); err != nil {
		return nil, err
	}

	if err := registerLocalExcreds(ctx, r, conn); err != nil {
		return nil, err
	}

	return r, nil
}

func registerLocalBlocks(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return blocks.RegisterBlocksServiceHandler(ctx, mux, conn)
}

func registerLocalOrders(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return orders.RegisterOrdersServiceHandler(ctx, mux, conn)
}

func registerLocalPassport(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return passport.RegisterPassportSeviceHandler(ctx, mux, conn)
}

func registerLocalTicks(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return ticks.RegisterHistoryServiceHandler(ctx, mux, conn)
}

func registerLocalUsers(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return users.RegisterUserServiceHandler(ctx, mux, conn)
}

func registerLocalStrategy(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return strategy.RegisterStrategyServiceHandler(ctx, mux, conn)
}

func registerLocalExcreds(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return excreds.RegisterExCredsServiceHandler(ctx, mux, conn)
}
