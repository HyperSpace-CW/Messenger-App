package repo

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"messanger/config"
)

func ConnectDB(cfg *config.Config) *sqlx.DB {
	db := sqlx.MustConnect("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.PG.Host,
		cfg.PG.Port,
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.DBName,
		cfg.PG.SSLMode,
	))

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	if err := Migration(db); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	return db
}

func Migration(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/repo/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
