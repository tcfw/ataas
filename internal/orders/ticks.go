package orders

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	ticksAPI "pm.tcfw.com.au/source/trader/api/pb/ticks"
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

		conn, err := grpc.Dial(userEndpoint, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		_ticksSvc = ticksAPI.NewHistoryServiceClient(conn)
	}

	return _ticksSvc, nil
}
