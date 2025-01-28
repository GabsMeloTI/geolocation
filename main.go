package main

import (
	"context"
	"geolocation/cmd"
	"geolocation/infra"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	defer stop()

	loadingEnv := infra.NewConfig()
	container := infra.NewContainerDI(loadingEnv)

	cmd.StartAPI(ctx, container)
}
