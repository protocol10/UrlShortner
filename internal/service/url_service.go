package service

import (
	"context"
	"fmt"

	"UrlShortner/internal/repository"
	"UrlShortner/internal/shortener"
)

type URLService interface {
	ShortenURL(ctx context.Context, longURL string) (string, error)
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
	// 1. Generate short code using the injected strategy
	shortCode, err := s.shortener.Shorten(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate short code: %w", err)
	}

	// 2. Save to database
	err = s.repo.Insert(ctx, longURL, shortCode)
	if err != nil {
		// If collision occurs, we could potentially retry with random generation or salt,
		// but for now we simply propagate the collision error as requested.
		if err == repository.ErrCollision {
			return "", fmt.Errorf("collision occurred for code %s: please try again", shortCode)
		}
		return "", fmt.Errorf("failed to save to database: %w", err)
	}

	return shortCode, nil
}
