package blocks

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	ordersAPI "pm.tcfw.com.au/source/trader/api/pb/orders"
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

		conn, err := grpc.Dial(ordersEndpoint, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		_ordersSvc = ordersAPI.NewOrdersServiceClient(conn)
	}

	return _ordersSvc, nil
}
