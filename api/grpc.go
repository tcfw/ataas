package api

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"pm.tcfw.com.au/source/trader/api/pb/blocks"
	"pm.tcfw.com.au/source/trader/api/pb/orders"
	"pm.tcfw.com.au/source/trader/api/pb/passport"
	"pm.tcfw.com.au/source/trader/api/pb/strategy"
	"pm.tcfw.com.au/source/trader/api/pb/ticks"
	"pm.tcfw.com.au/source/trader/api/pb/users"
	blocksImpl "pm.tcfw.com.au/source/trader/internal/blocks"
	ordersImpl "pm.tcfw.com.au/source/trader/internal/orders"
	passportImpl "pm.tcfw.com.au/source/trader/internal/passport"
	strategyImpl "pm.tcfw.com.au/source/trader/internal/strategies"
	ticksImpl "pm.tcfw.com.au/source/trader/internal/ticks"
	usersImpl "pm.tcfw.com.au/source/trader/internal/users"
)

func newGRPCServer(ctx context.Context, opts ...grpc.ServerOption) (func(), *grpc.Server) {
	grpcServer := grpc.NewServer(opts...)

	ticksServer, err := ticksImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	stratServer, err := strategyImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	blockServer, err := blocksImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	ordersServer, err := ordersImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	blocks.RegisterBlocksServiceServer(grpcServer, blockServer)
	orders.RegisterOrdersServiceServer(grpcServer, ordersServer)
	ticks.RegisterHistoryServiceServer(grpcServer, ticksServer)
	strategy.RegisterStrategyServiceServer(grpcServer, stratServer)
	users.RegisterUserServiceServer(grpcServer, &usersImpl.Server{})
	passport.RegisterPassportSeviceServer(grpcServer, &passportImpl.Server{})

	startServices := func() {
		blockServer.Listen()
		go ticksServer.Collect(ctx)
		go stratServer.Start(ctx)

		logrus.New().Infoln("Started services")
	}

	return startServices, grpcServer
}
