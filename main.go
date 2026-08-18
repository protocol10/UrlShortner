package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"UrlShortner/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Postgres Data Source Name (DSN) using the robust URL format
	dbConnectionStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	db, err := sql.Open("pgx", dbConnectionStr)
	if err != nil {
		log.Fatalf("Could not open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping database (check if running and credentials are correct): %v", err)
	}
	log.Println("Connected to PostgreSQL successfully")

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Close the *sql.DB connection since we only needed it for migrations
	db.Close()

	// ---------------------------------------------------------
	// Application Connection Pool Setup (pgxpool)
	// ---------------------------------------------------------
	poolConfig, err := pgxpool.ParseConfig(dbConnectionStr)
	if err != nil {
		log.Fatalf("Unable to parse database config for pool: %v", err)
	}

	// You can configure pool settings here (e.g. max connections)
	// poolConfig.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database with pgxpool: %v", err)
	}
	log.Println("pgx connection pool established successfully!")

	fmt.Println("URL Shortener initialized!")
	fmt.Printf("Using Algorithm: %s\n", cfg.Shortener.HashingAlgorithm)
	fmt.Printf("Max Character Limit: %d\n", cfg.Shortener.MaxCharLimit)
}
