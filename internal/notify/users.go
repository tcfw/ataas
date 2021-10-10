package notify

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	usersAPI "pm.tcfw.com.au/source/ataas/api/pb/users"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

var (
	_usersSvc usersAPI.UserServiceClient
)

func usersSvc() (usersAPI.UserServiceClient, error) {
	if _usersSvc == nil {
		usersEndpoint, envExists := os.LookupEnv("USERS_HOST")
		if !envExists {
			usersEndpoint = viper.GetString("grpc.addr")
		}

		conn, err := grpc.Dial(usersEndpoint, rpcUtils.InternalClientOptions()...)
		if err != nil {
			return nil, err
		}

		_usersSvc = usersAPI.NewUserServiceClient(conn)
	}

	return _usersSvc, nil
}
