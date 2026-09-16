package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func Test_postgresURLRepository_Insert(t *testing.T) {
	// Create a new mock pool
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := NewPostgresURLRepository(mock)

	tests := []struct {
		name      string
		longURL   string
		shortCode string
		mockSetup func()
		wantErr   error
	}{
		{
			name:      "Success",
			longURL:   "https://www.google.com",
			shortCode: "googl1",
			mockSetup: func() {
				mock.ExpectExec("INSERT INTO url_shortener").
					WithArgs("https://www.google.com", "googl1").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: nil,
		},
		{
			name:      "Collision",
			longURL:   "https://www.facebook.com",
			shortCode: "fb1234",
			mockSetup: func() {
				mock.ExpectExec("INSERT INTO url_shortener").
					WithArgs("https://www.facebook.com", "fb1234").
					WillReturnError(&pgconn.PgError{Code: "23505"})
			},
			wantErr: ErrCollision,
		},
		{
			name:      "Other DB Error",
			longURL:   "https://www.example.com",
			shortCode: "ex1234",
			mockSetup: func() {
				mock.ExpectExec("INSERT INTO url_shortener").
					WithArgs("https://www.example.com", "ex1234").
					WillReturnError(errors.New("some other error"))
			},
			wantErr: errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.Insert(context.Background(), tt.longURL, tt.shortCode)
			
			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.wantErr == ErrCollision {
					assert.Equal(t, ErrCollision, err)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
