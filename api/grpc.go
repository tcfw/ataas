package api

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"pm.tcfw.com.au/source/ataas/api/pb/blocks"
	"pm.tcfw.com.au/source/ataas/api/pb/excreds"
	"pm.tcfw.com.au/source/ataas/api/pb/notify"
	"pm.tcfw.com.au/source/ataas/api/pb/orders"
	"pm.tcfw.com.au/source/ataas/api/pb/passport"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/api/pb/users"
	blocksImpl "pm.tcfw.com.au/source/ataas/internal/blocks"
	excredsImpl "pm.tcfw.com.au/source/ataas/internal/excreds"
	notifyImpl "pm.tcfw.com.au/source/ataas/internal/notify"
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

	notifyServer, err := notifyImpl.NewServer(ctx)
	if err != nil {
		panic(err)
	}

	blocks.RegisterBlocksServiceServer(grpcServer, blockServer)
	excreds.RegisterExCredsServiceServer(grpcServer, excredsServer)
	notify.RegisterNotifyServiceServer(grpcServer, notifyServer)
	orders.RegisterOrdersServiceServer(grpcServer, ordersServer)
	passport.RegisterPassportSeviceServer(grpcServer, passportServer)
	strategy.RegisterStrategyServiceServer(grpcServer, stratServer)
	ticks.RegisterHistoryServiceServer(grpcServer, ticksServer)
	users.RegisterUserServiceServer(grpcServer, usersServer)

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
