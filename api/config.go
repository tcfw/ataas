package api

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("grpc.addr", ":19283")
	viper.SetDefault("gw.https.addr", ":8443")
	viper.SetDefault("gw.https.cert", "tls.cert")
	viper.SetDefault("gw.https.key", "tls.key")

	viper.SetDefault("gw.enableAuth", true)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	viper.SetDefault("collector.library", fmt.Sprintf("%s/trades", dir))
}
