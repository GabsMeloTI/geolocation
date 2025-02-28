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

	driver := e.Group("/driver")
	driver.POST("/create", container.HandlerDriver.CreateDriverHandler)
	driver.PUT("/update", container.HandlerDriver.UpdateDriverHandler)
	driver.PUT("/delete/:id", container.HandlerDriver.DeleteDriversHandler)

	announcement := e.Group("/announcement")
	announcement.POST("/create", container.HandlerAnnouncement.CreateAnnouncementHandler)
	announcement.PUT("/update", container.HandlerAnnouncement.UpdateAnnouncementHandler)
	announcement.PUT("/delete/:id", container.HandlerAnnouncement.DeleteAnnouncementHandler)

	trailer := e.Group("/trailer")
	trailer.POST("/create", container.HandlerTrailer.CreateTrailerHandler)
	trailer.PUT("/update", container.HandlerTrailer.UpdateTrailerHandler)
	trailer.PUT("/delete/:id", container.HandlerTrailer.DeleteTrailerHandler)

	tractorUnit := e.Group("/tractor-unit")
	tractorUnit.POST("/create", container.HandlerTractorUnit.CreateTractorUnitHandler)
	tractorUnit.PUT("/update", container.HandlerTractorUnit.UpdateTractorUnitHandler)
	tractorUnit.PUT("/delete/:id", container.HandlerTractorUnit.DeleteTractorUnitHandler)

	public := e.Group("/public")
	public.GET("/:ip", container.HandlerHist.GetPublicToken)
	public.POST("/check-route-tolls", container.HandlerNewRoutes.CalculateRoutes, _midlleware.CheckPublicAuthorization)

	user := e.Group("/user", _midlleware.CheckUserAuthorization)
	user.PUT("/delete", container.UserHandler.DeleteUser)
	user.PUT("/update", container.UserHandler.UpdateUser)
	user.PUT("/create", container.UserHandler.CreateUser)
	user.POST("/login", container.UserHandler.UserLogin)

	e.POST("/check-route-tolls", container.HandlerNewRoutes.CalculateRoutes, _midlleware.CheckAuthorization)
	e.POST("/google-route-tolls-public", container.HandlerRoutes.CheckRouteTolls, _midlleware.CheckPublicAuthorization)
	e.POST("/google-route-tolls", container.HandlerRoutes.CheckRouteTolls)
	e.GET("/ws", container.WsHandler.HandleWs)

	appointment := e.Group("/appointment")
	appointment.POST("/create", container.HandlerAppointment.CreateAppointmentHandler)
	appointment.PUT("/update", container.HandlerAppointment.UpdateAppointmentHandler)
	appointment.PUT("/delete/:id", container.HandlerAppointment.DeleteAppointmentsHandler)
	appointment.GET("/:id", container.HandlerAppointment.GetAppointmentByUserIDHandler)

	e.Logger.Fatal(e.Start(container.Config.ServerPort))
}
