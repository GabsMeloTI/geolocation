package infra

import (
	"database/sql"
	"geolocation/infra/database"
	"geolocation/infra/database/db_postgresql"
	"geolocation/infra/token"
	"geolocation/internal/advertisement"
	"geolocation/internal/drivers"
	"geolocation/internal/hist"
	"geolocation/internal/login"
	new_routes "geolocation/internal/new_routes"
	"geolocation/internal/routes"
	"geolocation/internal/tractor_unit"
	"geolocation/internal/trailer"
	"geolocation/internal/user"
	"geolocation/internal/ws"
	"geolocation/pkg/sso"
)

type ContainerDI struct {
	Config                  Config
	ConnDB                  *sql.DB
	HandlerRoutes           *routes.Handler
	ServiceRoutes           *routes.Service
	RepositoryRoutes        *routes.Repository
	HandlerNewRoutes        *new_routes.Handler
	ServiceNewRoutes        *new_routes.Service
	HandlerHist             *hist.Handler
	ServiceHist             *hist.Service
	RepositoryHist          *hist.Repository
	HandlerDriver           *drivers.Handler
	ServiceDriver           *drivers.Service
	RepositoryDriver        *drivers.Repository
	HandlerTractorUnit      *tractor_unit.Handler
	ServiceTractorUnit      *tractor_unit.Service
	RepositoryTractorUnit   *tractor_unit.Repository
	HandlerTrailer          *trailer.Handler
	ServiceTrailer          *trailer.Service
	RepositoryTrailer       *trailer.Repository
	HandlerAdvertisement    *advertisement.Handler
	ServiceAdvertisement    *advertisement.Service
	RepositoryAdvertisement *advertisement.Repository
	UserHandler             *user.Handler
	UserService             *user.Service
	UserRepository          *user.Repository
	WsHandler               *ws.Handler
	LoginHandler            *login.Handler
	LoginService            *login.Service
	LoginRepository         *login.Repository
	GoogleToken             *sso.GoogleToken
	PasetoMaker             *token.Maker
}

func NewContainerDI(config Config) *ContainerDI {
	container := &ContainerDI{Config: config}
	container.db()
	container.buildPkg()
	container.buildRepository()
	container.buildService()
	container.buildHandler()
	return container
}

func (c *ContainerDI) db() {
	dbConfig := database.Config{
		Host:        c.Config.DBHost,
		Port:        c.Config.DBPort,
		User:        c.Config.DBUser,
		Password:    c.Config.DBPassword,
		Database:    c.Config.DBDatabase,
		SSLMode:     c.Config.DBSSLMode,
		Driver:      c.Config.DBDriver,
		Environment: c.Config.Environment,
	}
	c.ConnDB = db_postgresql.NewConnection(&dbConfig)
}

func (c *ContainerDI) buildPkg() {
	c.GoogleToken = sso.NewGoogleToken(c.Config.GoogleClientId)
	maker, _ := token.NewPasetoMaker(c.Config.SignatureToken)
	c.PasetoMaker = &maker
}

func (c *ContainerDI) buildRepository() {
	c.RepositoryRoutes = routes.NewTollsRepository(c.ConnDB)
	c.RepositoryHist = hist.NewHistRepository(c.ConnDB)
	c.RepositoryDriver = drivers.NewDriversRepository(c.ConnDB)
	c.RepositoryTractorUnit = tractor_unit.NewTractorUnitsRepository(c.ConnDB)
	c.RepositoryTrailer = trailer.NewTrailersRepository(c.ConnDB)
	c.RepositoryAdvertisement = advertisement.NewAdvertisementsRepository(c.ConnDB)
	c.UserRepository = user.NewUserRepository(c.ConnDB)
	c.LoginRepository = login.NewRepository(c.ConnDB)
}

func (c *ContainerDI) buildService() {
	c.ServiceRoutes = routes.NewRoutesService(c.RepositoryRoutes, c.Config.GoogleMapsKey)
	c.ServiceNewRoutes = new_routes.NewRoutesNewService(c.RepositoryRoutes, c.Config.GoogleMapsKey)
	c.ServiceHist = hist.NewHistService(c.RepositoryHist, c.Config.SignatureToken)
	c.ServiceDriver = drivers.NewDriversService(c.RepositoryDriver)
	c.ServiceTractorUnit = tractor_unit.NewTractorUnitsService(c.RepositoryTractorUnit)
	c.ServiceTrailer = trailer.NewTrailersService(c.RepositoryTrailer)
	c.ServiceAdvertisement = advertisement.NewAdvertisementsService(c.RepositoryAdvertisement)
	c.UserService = user.NewUserService(c.UserRepository, c.Config.SignatureToken)
	c.LoginService = login.NewService(c.GoogleToken, c.LoginRepository, *c.PasetoMaker, c.Config.GoogleClientId)
}

func (c *ContainerDI) buildHandler() {
	c.HandlerRoutes = routes.NewRoutesHandler(c.ServiceRoutes)
	c.HandlerNewRoutes = new_routes.NewRoutesNewHandler(c.ServiceNewRoutes)
	c.HandlerHist = hist.NewHistHandler(c.ServiceHist)
	c.HandlerDriver = drivers.NewDriversHandler(c.ServiceDriver)
	c.HandlerTractorUnit = tractor_unit.NewTractorUnitsHandler(c.ServiceTractorUnit)
	c.HandlerTrailer = trailer.NewTrailersHandler(c.ServiceTrailer)
	c.HandlerAdvertisement = advertisement.NewAdvertisementHandler(c.ServiceAdvertisement)
	c.UserHandler = user.NewUserHandler(c.UserService, c.Config.GoogleClientId)
	c.LoginHandler = login.NewHandler(c.LoginService)
	hub := ws.NewHub()
	c.WsHandler = ws.NewWsHandler(hub)
	go hub.Run()

}
