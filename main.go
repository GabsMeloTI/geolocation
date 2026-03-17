package main

import (
	"context"
	"geolocation/pkg"
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
	pkg.InitRedis(loadingEnv.Environment, loadingEnv.SaveRedis)
	cmd.StartAPI(ctx, container)
}
