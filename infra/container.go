package infra

import (
	"database/sql"
	"geolocation/infra/database"
	"geolocation/infra/database/db_postgresql"
	"geolocation/internal/routes"
)

type ContainerDI struct {
	Config           Config
	ConnDB           *sql.DB
	HandlerRoutes    *routes.Handler
	ServiceRoutes    *routes.Service
	RepositoryRoutes *routes.Repository
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
}

func (c *ContainerDI) buildService() {
	c.ServiceRoutes = routes.NewRoutesService(c.RepositoryRoutes)
}

func (c *ContainerDI) buildHandler() {
	c.HandlerRoutes = routes.NewRoutesHandler(c.ServiceRoutes)
}
