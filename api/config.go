package api

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("grpc.addr", ":19283")
	viper.SetDefault("https.addr", ":8443")
	viper.SetDefault("tls.cert", "tls.cert")
	viper.SetDefault("tls.key", "tls.key")
}
