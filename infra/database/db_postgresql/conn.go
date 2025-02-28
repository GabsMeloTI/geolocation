package db_postgresql

import (
	"database/sql"
	"errors"
	"geolocation/infra/database"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewConnection(config *database.Config) *sql.DB {
	driver := config.Driver
	dsn := config.Driver + "://" + config.User + ":" + config.Password + "@" +
		config.Host + ":" + config.Port + "/" + config.Database + config.SSLMode
	db, err := sql.Open(driver, dsn)
	if err != nil {
		errConnection(config.Environment, err)
	}

	if err := runMigrations(db); err != nil {
		errConnection(config.Environment, err)
	}

	return db
}

func errConnection(environment string, err error) {
	panic("failed to connect " + environment + " postgres database_infra: " + err.Error())
}

func runMigrations(conn *sql.DB) error {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	migrationsPath := filepath.Join(pwd, "db/migration")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		panic(err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("Failed to run migrations:", err)
		return err
	}

	return nil
}
