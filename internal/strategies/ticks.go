package strategies

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	ticksAPI "pm.tcfw.com.au/source/ataas/api/pb/ticks"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

var (
	_ticksSvc ticksAPI.HistoryServiceClient
)

func ticksSvc() (ticksAPI.HistoryServiceClient, error) {
	if _ticksSvc == nil {
		ticksEndpoint, envExists := os.LookupEnv("TICKS_HOST")
		if !envExists {
			ticksEndpoint = viper.GetString("grpc.addr")
		}

		conn, err := grpc.Dial(ticksEndpoint, rpcUtils.InternalClientOptions()...)
		if err != nil {
			return nil, err
		}

		_ticksSvc = ticksAPI.NewHistoryServiceClient(conn)
	}

	return _ticksSvc, nil
}
