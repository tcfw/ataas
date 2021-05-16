package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"pm.tcfw.com.au/source/trader/api"
	"pm.tcfw.com.au/source/trader/db"
)

func init() {
	viper.SetConfigName("trader")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("trader")
	viper.AutomaticEnv()

	viper.SetDefault("db.url", "postgres://root@localhost:26257/trader_ticks?pool_max_conns=20")
	viper.SetDefault("services.start", true)
}

const (
	appVersion = "v0.0.0"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logrus.New()
	log.Infof("Auto Trader %s", appVersion)

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Warnln("No config found, proceeding with defaults/env")
		} else {
			log.Fatalf("Failed to init config: %s", err)
		}
	}

	if err := db.Init(ctx, viper.GetString("db.url")); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	apiServer, err := api.NewAPIServer(ctx)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := apiServer.Serve()
		if err != nil {
			log.Errorf("%s\n", err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Warnln("Shutting down...")
		apiServer.Stop()
		log.Infoln("Gracefully shutdown")
		os.Exit(0)
	}()

	wg.Wait()
}
