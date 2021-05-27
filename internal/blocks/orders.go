package blocks

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	ordersAPI "pm.tcfw.com.au/source/ataas/api/pb/orders"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

var (
	_ordersSvc ordersAPI.OrdersServiceClient
)

func ordersSvc() (ordersAPI.OrdersServiceClient, error) {
	if _ordersSvc == nil {
		ordersEndpoint, envExists := os.LookupEnv("ORDERS_HOST")
		if !envExists {
			ordersEndpoint = viper.GetString("grpc.addr")
		}

		conn, err := grpc.Dial(ordersEndpoint, rpcUtils.InternalClientOptions()...)
		if err != nil {
			return nil, err
		}

		_ordersSvc = ordersAPI.NewOrdersServiceClient(conn)
	}

	return _ordersSvc, nil
}
