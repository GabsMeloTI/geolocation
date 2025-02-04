package cmd

import (
	"context"
	"geolocation/infra"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
)

// @title GO-auth-service
// @description Document API
// @version 1.0
// @schemes https http
// @contact.name API Support
// @contact.url https://simpplify.com.br/contact
// @contact.email support@swagger.io
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func StartAPI(ctx context.Context, container *infra.ContainerDI) {
	e := echo.New()

	go func() {
		for {
			select {
			case <-ctx.Done():
				if err := e.Shutdown(ctx); err != nil {
					panic(err)
				}
				return
			default:
				time.Sleep(1 * time.Second)
			}
		}
	}()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: middleware.DefaultCORSConfig.AllowMethods,
	}))

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.POST("/check-route-tolls", container.HandlerRoutes.CheckRouteTolls)
	e.PUT("/added-favorite", container.HandlerRoutes.AddSavedRoutesFavorite)

	e.Logger.Fatal(e.Start(container.Config.ServerPort))
}
