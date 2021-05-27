package orders

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	excredsAPI "pm.tcfw.com.au/source/ataas/api/pb/excreds"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

var (
	_excredsSvc excredsAPI.ExCredsServiceClient
)

func excredsSvc() (excredsAPI.ExCredsServiceClient, error) {
	if _excredsSvc == nil {
		excredsEndpoint, envExists := os.LookupEnv("EXCREDS_HOST")
		if !envExists {
			excredsEndpoint = viper.GetString("grpc.addr")
		}

		conn, err := grpc.Dial(excredsEndpoint, rpcUtils.InternalClientOptions()...)
		if err != nil {
			return nil, err
		}

		_excredsSvc = excredsAPI.NewExCredsServiceClient(conn)
	}

	return _excredsSvc, nil
}
