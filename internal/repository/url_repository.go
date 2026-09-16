package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrCollision = errors.New("short code collision detected")
	ErrNotFound  = errors.New("url not found")
)

type URLRepository interface {
	Insert(ctx context.Context, longURL, shortCode string) error
	GetByShortCode(ctx context.Context, shortCode string) (string, error)
}

type DBExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type postgresURLRepository struct {
	db DBExecer
}

func NewPostgresURLRepository(db DBExecer) URLRepository {
	return &postgresURLRepository{db: db}
}

func (r *postgresURLRepository) Insert(ctx context.Context, longURL, shortCode string) error {
	query := `
		INSERT INTO url_shortener (url, short_code)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(ctx, query, longURL, shortCode)
	if err != nil {
		// Check for PostgreSQL unique constraint violation (code 23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrCollision
		}
		return err
	}
	return nil
}

func (r *postgresURLRepository) GetByShortCode(ctx context.Context, shortCode string) (string, error) {
	query := `
		SELECT url
		FROM url_shortener
		WHERE short_code = $1
	`
	var longURL string
	err := r.db.QueryRow(ctx, query, shortCode).Scan(&longURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return longURL, nil
}
