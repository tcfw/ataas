package blocks

import (
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	notifyAPI "pm.tcfw.com.au/source/ataas/api/pb/notify"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

var (
	_notifySvc notifyAPI.NotifyServiceClient
)

func notifySvc() (notifyAPI.NotifyServiceClient, error) {
	if _notifySvc == nil {
		notifyEndpoint, envExists := os.LookupEnv("NOTIFY_HOST")
		if !envExists {
			notifyEndpoint = viper.GetString("grpc.addr")
		}

		conn, err := grpc.Dial(notifyEndpoint, rpcUtils.InternalClientOptions()...)
		if err != nil {
			return nil, err
		}

		_notifySvc = notifyAPI.NewNotifyServiceClient(conn)
	}

	return _notifySvc, nil
}
