package infra

import (
	"database/sql"
	"geolocation/infra/database"
	"geolocation/infra/database/db_postgresql"
	"geolocation/internal/hist"
	new_routes "geolocation/internal/new_routes"
	"geolocation/internal/routes"
	"geolocation/internal/user"
)

type ContainerDI struct {
	Config           Config
	ConnDB           *sql.DB
	HandlerRoutes    *routes.Handler
	ServiceRoutes    *routes.Service
	RepositoryRoutes *routes.Repository
	HandlerNewRoutes *new_routes.Handler
	ServiceNewRoutes *new_routes.Service
	HandlerHist      *hist.Handler
	ServiceHist      *hist.Service
	RepositoryHist   *hist.Repository
	UserHandler      *user.Handler
	UserService      *user.Service
	UserRepository   *user.Repository
}

func NewContainerDI(config Config) *ContainerDI {
	container := &ContainerDI{Config: config}
	container.db()
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

func (c *ContainerDI) buildRepository() {
	c.RepositoryRoutes = routes.NewTollsRepository(c.ConnDB)
	c.RepositoryHist = hist.NewHistRepository(c.ConnDB)
	c.UserRepository = user.NewUserRepository(c.ConnDB)
}

func (c *ContainerDI) buildService() {
	c.ServiceRoutes = routes.NewRoutesService(c.RepositoryRoutes, c.Config.GoogleMapsKey)
	c.ServiceNewRoutes = new_routes.NewRoutesNewService(c.RepositoryRoutes, c.Config.GoogleMapsKey)
	c.ServiceHist = hist.NewHistService(c.RepositoryHist, c.Config.SignatureToken)
	c.UserService = user.NewUserService(c.UserRepository, c.Config.SignatureToken)
}

func (c *ContainerDI) buildHandler() {
	c.HandlerRoutes = routes.NewRoutesHandler(c.ServiceRoutes)
	c.HandlerNewRoutes = new_routes.NewRoutesNewHandler(c.ServiceNewRoutes)
	c.HandlerHist = hist.NewHistHandler(c.ServiceHist)
	c.UserHandler = user.NewUserHandler(c.UserService, c.Config.GoogleClientId)
}
