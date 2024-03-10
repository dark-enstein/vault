package main

import (
	"context"
	"encoding/gob"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/dark-enstein/vault/service"
	"github.com/spf13/pflag"
)

type InitConfig struct {
	debug bool
}

func init() {
	gob.Register(map[string]string{})
}

func main() {
	var i = InitConfig{}
	pflag.BoolVarP(&i.debug, "debug", "d", true, "enable debug mode")
	pflag.Parse()

	// starting
	ctx := context.Background()
	logger := vlog.New(i.debug)
	srv, err := service.New(ctx, logger, service.WithStoreStr(service.STORE_FILE), service.WithFileLoc("./gob"))
	if err != nil {
		logger.Logger().Fatal().Msgf("error while setting up service: %w", err)
	}
	if err = srv.Run(ctx); err != nil {
		logger.Logger().Fatal().Msgf("error while service is starting: %s\n", err.Error())
	}
}
