package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrCollision = errors.New("short code collision detected")
)

type URLRepository interface {
	Insert(ctx context.Context, longURL, shortCode string) error
}

type DBExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
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
