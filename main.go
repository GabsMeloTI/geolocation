package main

import (
	"context"
	"os/signal"
	"syscall"

	"geolocation/cmd"
	_ "geolocation/docs"
	"geolocation/infra"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
	)
	defer stop()

	loadingEnv := infra.NewConfig()
	container := infra.NewContainerDI(loadingEnv)
	// pkg.InitRedis(loadingEnv.Environment)
	cmd.StartAPI(ctx, container)
}
