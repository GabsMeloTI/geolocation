package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "geolocation/docs"
	_ "geolocation/docs/limited"

	"geolocation/infra"
	_midlleware "geolocation/infra/middleware"
)

// @title Geolocation
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

	// Configurar JSON serializer personalizado para não escapar barras
	e.JSONSerializer = &CustomJSONSerializer{}

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

	e.GET("/swagger/3744f3e2-4989-433b-83e9-b467f5c7d503/*", echoSwagger.WrapHandler)
	e.GET("/swagger/system/documentation/*", echoSwagger.EchoWrapHandler(echoSwagger.InstanceName("limited")))

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	driver := e.Group("/driver", _midlleware.CheckUserAuthorization)
	driver.POST("/create", container.HandlerDriver.CreateDriverHandler)
	driver.PUT("/update", container.HandlerDriver.UpdateDriverHandler)
	driver.PUT("/delete/:id", container.HandlerDriver.DeleteDriversHandler)
	driver.GET("/list", container.HandlerDriver.GetDriverHandler)
	driver.GET("/list/:id", container.HandlerDriver.GetDriverByIdHandler)

	advertisement := e.Group("/advertisement", _midlleware.CheckUserAuthorization)
	advertisement.POST("/create", container.HandlerAdvertisement.CreateAdvertisementHandler)
	advertisement.POST("/finish/create", container.HandlerAdvertisement.UpdatedAdvertisementFinishedCreate)
	advertisement.PUT("/update", container.HandlerAdvertisement.UpdateAdvertisementHandler)
	advertisement.PUT("/delete/:id", container.HandlerAdvertisement.DeleteAdvertisementHandler)
	advertisement.GET("/list", container.HandlerAdvertisement.GetAllAdvertisementHandler)
	advertisement.GET("/list/:id", container.HandlerAdvertisement.GetAdvertisementByIDService)
	advertisement.GET("/list/by-user", container.HandlerAdvertisement.GetAllAdvertisementByUserHandler)
	advertisement.PUT("/update/route", container.HandlerAdvertisement.UpdateAdsRouteChoose)

	trailer := e.Group("/trailer", _midlleware.CheckUserAuthorization)
	trailer.POST("/create", container.HandlerTrailer.CreateTrailerHandler)
	trailer.PUT("/update", container.HandlerTrailer.UpdateTrailerHandler)
	trailer.PUT("/delete/:id", container.HandlerTrailer.DeleteTrailerHandler)
	trailer.GET("/list", container.HandlerTrailer.GetTrailerHandler)
	trailer.GET("/list/:id", container.HandlerTrailer.GetTrailerByIdHandler)

	tractorUnit := e.Group("/tractor-unit", _midlleware.CheckUserAuthorization)
	tractorUnit.POST("/create", container.HandlerTractorUnit.CreateTractorUnitHandler)
	tractorUnit.PUT("/update", container.HandlerTractorUnit.UpdateTractorUnitHandler)
	tractorUnit.PUT("/delete/:id", container.HandlerTractorUnit.DeleteTractorUnitHandler)
	tractorUnit.GET("/list", container.HandlerTractorUnit.GetTractorUnitHandler)
	tractorUnit.GET("/list/:id", container.HandlerTractorUnit.GetTractorUnitByIdHandler)

	attach := e.Group("/attach", _midlleware.CheckUserAuthorization)
	attach.POST("/upload", container.HandlerAttachment.CreateAttachHandler)
	attach.PUT("/update", container.HandlerAttachment.UpdateAttachHandler)
	attach.GET("/list/:type", container.HandlerAttachment.GetAllAttachmentById)

	zonasRisco := e.Group("/zonas-risco")
	zonasRisco.POST("/create", container.HandlerZonasRisco.CreateZonaRiscoHandler)
	zonasRisco.PUT("/update", container.HandlerZonasRisco.UpdateZonaRiscoHandler)
	zonasRisco.PUT("/delete/:id", container.HandlerZonasRisco.DeleteZonaRiscoHandler)
	zonasRisco.GET("/list/all/:id", container.HandlerZonasRisco.GetAllZonasRiscoHandler)
	zonasRisco.GET("/list/:id", container.HandlerZonasRisco.GetZonaRiscoByIdHandler)

	e.POST("/recover-password", container.UserHandler.RecoverPassword)
	e.PUT("/recover-password/confirm", container.UserHandler.ConfirmRecoverPassword, _midlleware.CheckUserAuthorization)

	user := e.Group("/user", _midlleware.CheckUserAuthorization)
	user.GET("/info", container.UserHandler.GetUserInfo)
	user.PUT("/delete", container.UserHandler.DeleteUser)
	user.PUT("/update", container.UserHandler.UpdateUser)
	user.PUT("/update/password", container.UserHandler.UpdateUserPassword)
	user.PUT("/address/update", container.UserHandler.UpdateUserAddress)
	user.PUT("/personal/update", container.UserHandler.UpdateUserPersonalInfo)
	user.POST("/plan", container.HandlerUserPlan.CreateUserPlanHandler)
	e.GET("/user/email", container.UserHandler.UserExists)
	e.POST("/caminhao/carbono", container.UserHandler.InfoCaminhao)
	e.GET("/consulta/:placa", container.UserHandler.ConsultarPlaca)
	e.POST("/consulta/multiplas", container.UserHandler.ConsultarMultiplasPlacas)

	public := e.Group("/public")
	public.GET("/:ip", container.HandlerHist.GetPublicToken)
	public.GET("/advertisement/list", container.HandlerAdvertisement.GetAllAdvertisementPublicHandler)
	public.GET("/advertisement/list/:id", container.HandlerAdvertisement.GetAdvertisementByIDPublicService)
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

	// simpplify
	e.POST("/check-route-tolls-simpplify", container.HandlerNewRoutes.CalculateRoutes, _midlleware.CheckAuthorization)
	e.POST("/check-route-tolls-simpplify-cep", container.HandlerNewRoutes.CalculateRoutesWithCEP, _midlleware.CheckAuthorization)
	e.POST("/route-cep", container.HandlerNewRoutes.CalculateRoutesCEP)
	e.POST("/route-cep-avoidance", container.HandlerNewRoutes.CalculateDistancesBetweenPointsWithRiskAvoidanceHandler)
	e.POST("/route-coordinate-avoidance", container.HandlerNewRoutes.CalculateDistancesBetweenPointsWithRiskAvoidanceFromCoordinatesHandler)
	e.POST("/nearby-location", container.HandlerNewRoutes.CalculateDistancesFromOrigin)

	// easyfrete
	e.POST("/check-route-tolls-easy", container.HandlerNewRoutes.CalculateRoutes, _midlleware.CheckUserAuthorization)
	e.POST("/check-route-tolls-coordinate", container.HandlerNewRoutes.CalculateRoutesWithCoordinate, _midlleware.CheckUserAuthorization)
	e.POST("/check-route-tolls-cep", container.HandlerNewRoutes.CalculateRoutesWithCEP, _midlleware.CheckUserAuthorization)
	e.POST("/google-route-tolls-public", container.HandlerRoutes.CheckRouteTolls, _midlleware.CheckPublicAuthorization)
	e.POST("/google-route-tolls", container.HandlerRoutes.CheckRouteTolls)
	e.POST("/login", container.LoginHandler.Login)
	e.POST("/create", container.LoginHandler.CreateUser)
	e.POST("/create/client", container.LoginHandler.CreateUserClient, _midlleware.CheckUserAuthorization)
	e.POST("/webhook/stripe", container.HandlerPayment.StripeWebhookHandler)
	e.GET("/ws", container.WsHandler.HandleWs, _midlleware.CheckUserWsAuthorization)
	e.GET("/token", container.HandlerUserPlan.GetTokenUserHandler, _midlleware.CheckUserAuthorization)
	e.GET("/dashboard", container.HandlerDashboard.GetDashboardHandler, _midlleware.CheckUserAuthorization)
	e.GET("/check/:plate", container.HandlerTractorUnit.CheckPlateHandler)
	e.GET("/payment-history", container.HandlerPayment.GetPaymentHistHandler, _midlleware.CheckUserAuthorization)

	appointment := e.Group("/appointment")
	appointment.PUT("/update", container.HandlerAppointment.UpdateAppointmentHandler)
	appointment.PUT("/delete/:id", container.HandlerAppointment.DeleteAppointmentsHandler)
	appointment.GET("/:id", container.HandlerAppointment.GetAppointmentByUserIDHandler)

	address := e.Group("/address")
	address.GET("/find", container.HandlerAddress.FindAddressByQueryHandler)
	address.GET("/find/v2", container.HandlerAddress.FindAddressByQueryV2Handler)
	address.GET("/find/cep/v2", container.HandlerAddress.FindUniqueAddressesByCEPHandler)
	address.GET("/find/:cep", container.HandlerAddress.FindAddressByCEPHandler)
	address.GET("/state", container.HandlerAddress.FindStateAll)
	address.GET("/city/:idState", container.HandlerAddress.FindCityAll)

	locations := e.Group("/locations", _midlleware.CheckAuthorization)
	locations.POST("/create", container.HandlerLocation.CreateLocationHandler)
	locations.PUT("/update", container.HandlerLocation.UpdateLocationHandler)
	locations.DELETE("/delete/:id", container.HandlerLocation.DeleteLocationHandler)
	locations.GET("/list/:providerId", container.HandlerLocation.GetLocationHandler)

	e.Logger.Fatal(e.Start(container.Config.ServerPort))
}

// CustomJSONSerializer é um serializer JSON personalizado que não escapa barras
type CustomJSONSerializer struct{}

func (j *CustomJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	enc.SetEscapeHTML(false)
	return enc.Encode(i)
}

func (j *CustomJSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	return json.NewDecoder(c.Request().Body).Decode(i)
}
