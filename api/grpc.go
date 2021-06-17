package api

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"pm.tcfw.com.au/source/ataas/api/pb/blocks"
	"pm.tcfw.com.au/source/ataas/api/pb/excreds"
	"pm.tcfw.com.au/source/ataas/api/pb/orders"
	"pm.tcfw.com.au/source/ataas/api/pb/passport"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/api/pb/users"
	blocksImpl "pm.tcfw.com.au/source/ataas/internal/blocks"
	excredsImpl "pm.tcfw.com.au/source/ataas/internal/excreds"
	ordersImpl "pm.tcfw.com.au/source/ataas/internal/orders"
	passportImpl "pm.tcfw.com.au/source/ataas/internal/passport"
	strategyImpl "pm.tcfw.com.au/source/ataas/internal/strategies"
	ticksImpl "pm.tcfw.com.au/source/ataas/internal/ticks"
	usersImpl "pm.tcfw.com.au/source/ataas/internal/users"
)

func newGRPCServer(ctx context.Context, opts ...grpc.ServerOption) (func(), func(), *grpc.Server) {
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

	usersServer, err := usersImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	passportServer, err := passportImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	excredsServer, err := excredsImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	blocks.RegisterBlocksServiceServer(grpcServer, blockServer)
	orders.RegisterOrdersServiceServer(grpcServer, ordersServer)
	ticks.RegisterHistoryServiceServer(grpcServer, ticksServer)
	strategy.RegisterStrategyServiceServer(grpcServer, stratServer)
	users.RegisterUserServiceServer(grpcServer, usersServer)
	passport.RegisterPassportSeviceServer(grpcServer, passportServer)
	excreds.RegisterExCredsServiceServer(grpcServer, excredsServer)

	go ticksServer.Collect(ctx)
	startServices := func() {
		blockServer.Listen()
		go stratServer.Start(ctx)

		logrus.New().Infoln("Started services")
	}

	stopServices := func() {
		blockServer.Stop()

		err := ticksServer.Close()
		if err != nil {
			panic(err)
		}

	}

	return startServices, stopServices, grpcServer
}
