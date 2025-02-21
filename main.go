package main

import (
	"context"
	"geolocation/cmd"
	"geolocation/infra"
	"geolocation/pkg"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	defer stop()

	loadingEnv := infra.NewConfig()
	container := infra.NewContainerDI(loadingEnv)

	pkg.InitRedis()
	cmd.StartAPI(ctx, container)
}

//testing
