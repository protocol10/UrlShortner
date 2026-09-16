package service

import (
	"context"
	"fmt"
	"time"

	"UrlShortner/internal/repository"
	"UrlShortner/internal/shortener"
)

type URLService interface {
	ShortenURL(ctx context.Context, longURL string) (string, error)
	GetLongURL(ctx context.Context, shortCode string) (string, error)
}

type urlService struct {
	shortener strategy
	repo      repository.URLRepository
}

// using alias to avoid stuttering with package name
type strategy = shortener.Shortener

func NewURLService(s strategy, repo repository.URLRepository) URLService {
	return &urlService{
		shortener: s,
		repo:      repo,
	}
}

func (s *urlService) ShortenURL(ctx context.Context, longURL string) (string, error) {
	const maxRetries = 3
	var shortCode string
	var err error

	for i := 0; i < maxRetries; i++ {
		// Since we don't want deduplication, we ALWAYS append a timestamp salt.
		// We also append 'i' so if a rare collision happens, the salt changes on retry.
		hashInput := fmt.Sprintf("%s-%d-%d", longURL, time.Now().UnixNano(), i)

		// 1. Generate short code using the injected strategy
		shortCode, err = s.shortener.Shorten(hashInput)
		if err != nil {
			return "", fmt.Errorf("failed to generate short code: %w", err)
		}

		// 2. Save to database (Always save the ORIGINAL longURL, not the salted one!)
		err = s.repo.Insert(ctx, longURL, shortCode)
		if err == nil {
			return shortCode, nil // Success!
		}

		// If the error is anything OTHER than a collision, fail immediately
		if err != repository.ErrCollision {
			return "", fmt.Errorf("failed to save to database: %w", err)
		}
		
		// If it WAS a collision, the loop will restart and try again with a new salt!
	}

	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxRetries)
}

func (s *urlService) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	if shortCode == "" {
		return "", fmt.Errorf("short code is required")
	}
	return s.repo.GetByShortCode(ctx, shortCode)
}
