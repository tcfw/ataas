package orders

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
		userEndpoint, envExists := os.LookupEnv("TICKS_HOST")
		if !envExists {
			userEndpoint = viper.GetString("grpc.addr")
		}

		conn, err := grpc.Dial(userEndpoint, rpcUtils.InternalClientOptions()...)
		if err != nil {
			return nil, err
		}

		_ticksSvc = ticksAPI.NewHistoryServiceClient(conn)
	}

	return _ticksSvc, nil
}
