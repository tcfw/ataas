package api

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("grpc.addr", ":19283")
	viper.SetDefault("https.addr", ":8443")
	viper.SetDefault("tls.cert", "tls.cert")
	viper.SetDefault("tls.key", "tls.key")

	viper.SetDefault("gw.enableAuth", true)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	viper.SetDefault("collector.library", fmt.Sprintf("%s/trades", dir))
}
