package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "geolocation/docs"
	"geolocation/infra"
	_midlleware "geolocation/infra/middleware"
	"geolocation/internal/webhook"
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
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
	}))

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	driver := e.Group("/driver", _midlleware.CheckUserAuthorization)
	driver.POST("/create", container.HandlerDriver.CreateDriverHandler)
	driver.PUT("/update", container.HandlerDriver.UpdateDriverHandler)
	driver.PUT("/delete/:id", container.HandlerDriver.DeleteDriversHandler)
	driver.GET("/list", container.HandlerDriver.GetDriverHandler)

	advertisement := e.Group("/advertisement", _midlleware.CheckUserAuthorization)
	advertisement.POST("/create", container.HandlerAdvertisement.CreateAdvertisementHandler)
	advertisement.POST(
		"/finish/create",
		container.HandlerAdvertisement.UpdatedAdvertisementFinishedCreate,
	)
	advertisement.PUT("/update", container.HandlerAdvertisement.UpdateAdvertisementHandler)
	advertisement.PUT("/delete/:id", container.HandlerAdvertisement.DeleteAdvertisementHandler)
	advertisement.GET("/list", container.HandlerAdvertisement.GetAllAdvertisementHandler)
	advertisement.GET(
		"/list/by-user",
		container.HandlerAdvertisement.GetAllAdvertisementByUserHandler,
	)
	advertisement.PUT("/update/route", container.HandlerAdvertisement.UpdateAdsRouteChoose)

	trailer := e.Group("/trailer", _midlleware.CheckUserAuthorization)
	trailer.POST("/create", container.HandlerTrailer.CreateTrailerHandler)
	trailer.PUT("/update", container.HandlerTrailer.UpdateTrailerHandler)
	trailer.PUT("/delete/:id", container.HandlerTrailer.DeleteTrailerHandler)
	trailer.GET("/list", container.HandlerTrailer.GetTrailerHandler)

	tractorUnit := e.Group("/tractor-unit", _midlleware.CheckUserAuthorization)
	tractorUnit.POST("/create", container.HandlerTractorUnit.CreateTractorUnitHandler)
	tractorUnit.PUT("/update", container.HandlerTractorUnit.UpdateTractorUnitHandler)
	tractorUnit.PUT("/delete/:id", container.HandlerTractorUnit.DeleteTractorUnitHandler)
	tractorUnit.GET("/list", container.HandlerTractorUnit.GetTractorUnitHandler)

	attach := e.Group("/attach", _midlleware.CheckUserAuthorization)
	attach.POST("/upload", container.HandlerAttachment.CreateAttachHandler)
	attach.PUT("/delete/:id", container.HandlerAttachment.DeleteAttachHandler)

	user := e.Group("/user", _midlleware.CheckUserAuthorization)
	user.GET("/info", container.UserHandler.GetUserInfo)
	user.PUT("/delete", container.UserHandler.DeleteUser)
	user.PUT("/update", container.UserHandler.UpdateUser)
	user.PUT("/address/update", container.UserHandler.UpdateUserAddress)
	user.PUT("/personal/update", container.UserHandler.UpdateUserPersonalInfo)
	user.POST("/plan", container.HandlerUserPlan.CreateUserPlanHandler)

	public := e.Group("/public")
	public.GET("/:ip", container.HandlerHist.GetPublicToken)
	public.GET("/advertisement/list", container.HandlerAdvertisement.GetAllAdvertisementPublicHandler)
	// easyfrete no user
	public.POST("/check-route-tolls", container.HandlerNewRoutes.CalculateRoutes, _midlleware.CheckPublicAuthorization)

	route := e.Group("/route", _midlleware.CheckUserAuthorization)
	route.GET("/favorite/list", container.HandlerNewRoutes.GetFavoriteRouteHandler)
	route.PUT("/favorite/remove/:id", container.HandlerNewRoutes.RemoveFavoriteRouteHandler)
	route.POST("/simple", container.HandlerNewRoutes.GetSimpleRoute)

	chat := e.Group("/chat", _midlleware.CheckUserAuthorization)
	chat.POST("/create-room", container.WsHandler.CreateChatRoom)
	chat.POST("/update-offer", container.WsHandler.UpdateMessageOffer)
	chat.GET("/messages/:room_id", container.WsHandler.GetMessagesByRoomId)
	chat.POST("/update-freight", container.WsHandler.UpdateFreightLocation)
	e.GET("/ws", container.WsHandler.HandleWs, _midlleware.CheckUserWsAuthorization)

	// simpplify
	e.POST(
		"/check-route-tolls-simpplify",
		container.HandlerNewRoutes.CalculateRoutes,
		_midlleware.CheckAuthorization,
	)
	// easyfrete
	e.POST(
		"/check-route-tolls-easy",
		container.HandlerNewRoutes.CalculateRoutes,
		_midlleware.CheckUserAuthorization,
	)
	e.POST(
		"/google-route-tolls-public",
		container.HandlerRoutes.CheckRouteTolls,
		_midlleware.CheckPublicAuthorization,
	)
	e.POST("/google-route-tolls", container.HandlerRoutes.CheckRouteTolls)
	e.POST("/login", container.LoginHandler.Login)
	e.POST("/create", container.LoginHandler.CreateUser)

	appointment := e.Group("/appointment")
	appointment.PUT("/update", container.HandlerAppointment.UpdateAppointmentHandler)
	appointment.PUT("/delete/:id", container.HandlerAppointment.DeleteAppointmentsHandler)
	appointment.GET("/:id", container.HandlerAppointment.GetAppointmentByUserIDHandler)

	e.POST("/webhook-payment", webhook.WebhookPaymentHandler)
	e.POST("/webhook/stripe", container.HandlerPayment.StripeWebhookHandler)
	e.GET(
		"/payment-history",
		container.HandlerPayment.GetPaymentHistHandler,
		_midlleware.CheckUserAuthorization,
	)

	address := e.Group("/address")
	address.GET("/find", container.HandlerAddress.FindAddressByQueryHandler)
	address.GET("/state", container.HandlerAddress.FindStateAll)
	address.GET("/city/:idState", container.HandlerAddress.FindCityAll)

	e.GET(
		"/token",
		container.HandlerUserPlan.GetTokenUserHandler,
		_midlleware.CheckUserAuthorization,
	)

	e.GET("/dashboard", container.HandlerDashboard.GetDashboardHandler, _midlleware.CheckUserAuthorization)

	certFile := "fullchain.pem"
	keyFile := "privkey.pem"

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("Certificado não encontrado: %v", err)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("Chave privada não encontrada: %v", err)
	}

	e.Logger.Fatal(e.StartTLS(container.Config.ServerPort, certFile, keyFile))
}
