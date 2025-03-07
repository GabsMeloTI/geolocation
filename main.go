package main

import (
	"context"
	"geolocation/cmd"
	_ "geolocation/docs"
	"geolocation/infra"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	defer stop()

	loadingEnv := infra.NewConfig()
	container := infra.NewContainerDI(loadingEnv)

	//pkg.InitRedis(loadingEnv.Environment)
	cmd.StartAPI(ctx, container)
}
