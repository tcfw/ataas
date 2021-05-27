package strategies

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	blocksAPI "pm.tcfw.com.au/source/ataas/api/pb/blocks"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

var (
	_blocksSvc blocksAPI.BlocksServiceClient
)

func blocksSvc() (blocksAPI.BlocksServiceClient, error) {
	if _blocksSvc == nil {
		blocksEndpoint, envExists := os.LookupEnv("BLOCKS_HOST")
		if !envExists {
			blocksEndpoint = viper.GetString("grpc.addr")
		}

		conn, err := grpc.Dial(blocksEndpoint, rpcUtils.InternalClientOptions()...)
		if err != nil {
			return nil, err
		}

		_blocksSvc = blocksAPI.NewBlocksServiceClient(conn)
	}

	return _blocksSvc, nil
}
