package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"UrlShortner/internal/handler"
	"UrlShortner/internal/repository"
	"UrlShortner/internal/service"
	"UrlShortner/internal/shortener"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAPI_ShortenURL(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up a Postgres container
	dbName := "urlshortener"
	dbUser := "postgres"
	dbPassword := "postgres"

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	// Ensure the container is cleaned up when the test finishes
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// 2. Run Migrations
	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)

	// Find absolute path to migrations folder
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsPath := filepath.Join(basepath, "..", "..", "migrations")

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)
	require.NoError(t, err)

	err = m.Up()
	require.NoError(t, err, "failed to run migrations")
	_ = db.Close() // close migration connection

	// 3. Initialize the Application
	poolConfig, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)
	defer pool.Close()

	shortenerStrategy, err := shortener.New(shortener.Config{
		Approach:         "hash",
		HashingAlgorithm: "xxhash",
		MaxCharLimit:     8,
	})
	require.NoError(t, err)

	// Wire it all up
	urlRepo := repository.NewPostgresURLRepository(pool)
	urlService := service.NewURLService(shortenerStrategy, urlRepo)
	urlHandler := handler.NewURLHandler(urlService)

	// Start HTTP test server
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/shorten", urlHandler.HandleShorten)
	mux.HandleFunc("GET /api/v1/{shortCode}", urlHandler.HandleGet)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 4. Test the API

	// Test Case 1: Shorten a valid URL
	originalURL := "https://www.github.com/protocol10"
	reqBody := map[string]string{
		"url": originalURL,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/api/v1/shorten", "application/json", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody struct {
		ShortCode string `json:"short_code"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.NotEmpty(t, respBody.ShortCode)

	// Test Case 2: Retrieve original URL using GET /{shortCode} (expects HTTP 302 Found redirect)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	getResp, err := client.Get(fmt.Sprintf("%s/api/v1/%s", ts.URL, respBody.ShortCode))
	require.NoError(t, err)
	defer func() { _ = getResp.Body.Close() }()

	assert.Equal(t, http.StatusFound, getResp.StatusCode)
	assert.Equal(t, originalURL, getResp.Header.Get("Location"))

	// Test Case 3: GET with unknown short code returns 404
	notFoundResp, err := http.Get(ts.URL + "/api/v1/unknown123")
	require.NoError(t, err)
	defer func() { _ = notFoundResp.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, notFoundResp.StatusCode)

	// Test Case 4: Missing URL on POST returns 400
	reqBody = map[string]string{}
	bodyBytes, _ = json.Marshal(reqBody)
	resp2, err := http.Post(ts.URL+"/api/v1/shorten", "application/json", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)
}

func TestAPI_CounterStrategy(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up Postgres container
	dbName := "urlshortener"
	dbUser := "postgres"
	dbPassword := "postgres"

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run Migrations
	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsPath := filepath.Join(basepath, "..", "..", "migrations")

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)
	require.NoError(t, err)

	err = m.Up()
	require.NoError(t, err)
	_ = db.Close()

	// 2. Spin up Redis container
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	require.NoError(t, err, "failed to start redis container")
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate redisContainer: %s", err)
		}
	}()

	redisEndpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: redisEndpoint,
	})

	// 3. Initialize App with "counter" strategy
	poolConfig, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)
	defer pool.Close()

	counterStrategy, err := shortener.New(shortener.Config{
		Approach:              "counter",
		MaxCharLimit:          6,
		CounterClient:         shortener.NewRedisCounterAdapter(rdb),
		CounterKey:            "test:counter",
		ObfuscationMultiplier: 11400714819323198485,
		ObfuscationXOR:        0x5BF03635467C061D,
	})
	require.NoError(t, err)

	urlRepo := repository.NewPostgresURLRepository(pool)
	urlService := service.NewURLService(counterStrategy, urlRepo)
	urlHandler := handler.NewURLHandler(urlService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/shorten", urlHandler.HandleShorten)
	mux.HandleFunc("GET /api/v1/{shortCode}", urlHandler.HandleGet)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 4. Shorten a URL via Counter Strategy
	originalURL := "https://www.redis.io"
	reqBody := map[string]string{"url": originalURL}
	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/api/v1/shorten", "application/json", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody struct {
		ShortCode string `json:"short_code"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.NotEmpty(t, respBody.ShortCode)

	// Verify redirect (GET) works
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	getResp, err := client.Get(fmt.Sprintf("%s/api/v1/%s", ts.URL, respBody.ShortCode))
	require.NoError(t, err)
	defer func() { _ = getResp.Body.Close() }()

	assert.Equal(t, http.StatusFound, getResp.StatusCode)
	assert.Equal(t, originalURL, getResp.Header.Get("Location"))
}
