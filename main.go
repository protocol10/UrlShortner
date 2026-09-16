package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"UrlShortner/config"
	"UrlShortner/internal/handler"
	"UrlShortner/internal/repository"
	"UrlShortner/internal/service"
	"UrlShortner/internal/shortener"

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

	// ---------------------------------------------------------
	// 1. Run Migrations
	// ---------------------------------------------------------
	dbConnectionStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	db, err := sql.Open("pgx", dbConnectionStr)
	if err != nil {
		log.Fatalf("Could not open database connection for migrations: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	db.Close() // Close migration connection

	// ---------------------------------------------------------
	// 2. Application Connection Pool Setup
	// ---------------------------------------------------------
	poolConfig, err := pgxpool.ParseConfig(dbConnectionStr)
	if err != nil {
		log.Fatalf("Unable to parse database config for pool: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database with pgxpool: %v", err)
	}
	log.Println("pgx connection pool established successfully!")

	// ---------------------------------------------------------
	// 3. Initialize Strategies and Layers
	// ---------------------------------------------------------
	// Create Strategy
	shortenerStrategy, err := shortener.New(shortener.Config{
		Approach:         cfg.Shortener.Approach,
		HashingAlgorithm: cfg.Shortener.HashingAlgorithm,
		MaxCharLimit:     cfg.Shortener.MaxCharLimit,
	})
	if err != nil {
		log.Fatalf("Failed to initialize shortener strategy: %v", err)
	}
	log.Printf("Shortener strategy initialized: Approach=%s, Algorithm=%s, Limit=%d",
		cfg.Shortener.Approach, cfg.Shortener.HashingAlgorithm, cfg.Shortener.MaxCharLimit)

	// Dependency Injection: Repo -> Service -> Handler
	urlRepo := repository.NewPostgresURLRepository(pool)
	urlService := service.NewURLService(shortenerStrategy, urlRepo)
	urlHandler := handler.NewURLHandler(urlService)

	// ---------------------------------------------------------
	// 4. Start HTTP Server
	// ---------------------------------------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/shorten", urlHandler.HandleShorten)
	mux.HandleFunc("GET /api/v1/{shortCode}", urlHandler.HandleGet)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
