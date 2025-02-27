package cmd

import (
	"context"
	"geolocation/infra"
	_midlleware "geolocation/infra/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
)

func StartAPI(ctx context.Context, container *infra.ContainerDI) {
	e := echo.New()

	// Graceful Shutdown
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
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods: middleware.DefaultCORSConfig.AllowMethods,
	}))

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.POST("/check-route-tolls", container.HandlerRoutes.CheckRouteTolls)
	e.POST("/route", container.HandlerNewRoutes.CalculateRoutes)
	e.POST("/check-route-tolls-public", container.HandlerRoutes.CheckRouteTolls, _midlleware.CheckPublicAuthorization)
	e.GET("/public/:ip", container.HandlerHist.GetPublicToken)

	e.POST("/create-user", container.UserHandler.CreateUser)
	e.POST("/login", container.UserHandler.UserLogin)

	user := e.Group("/user", _midlleware.CheckUserAuthorization)
	user.PUT("/delete", container.UserHandler.DeleteUser)
	user.PUT("/update", container.UserHandler.UpdateUser)

	e.GET("/ws", container.WsHandler.HandleWs)

	e.Logger.Fatal(e.Start(container.Config.ServerPort))
}
