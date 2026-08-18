package shortener

import (
	"fmt"
)

// Shortener is the master strategy interface.
type Shortener interface {
	Shorten(longURL string) (string, error)
}

// Config represents the parameters needed to initialize a shortener.
type Config struct {
	Approach         string // "hash" or "random"
	HashingAlgorithm string // "md5", "sha256", "fnv", "murmur", "xxhash"
	MaxCharLimit     int
}

// New creates and returns a Shortener based on the provided configuration.
func New(cfg Config) (Shortener, error) {
	if cfg.Approach == "random" {
		return NewRandomShortener(cfg.MaxCharLimit), nil
	}

	if cfg.Approach == "hash" {
		var h Hasher
		switch cfg.HashingAlgorithm {
		case "md5":
			h = &MD5Hasher{}
		case "sha256":
			h = &SHA256Hasher{}
		case "fnv":
			h = &FNVHasher{}
		case "murmur":
			h = &MurmurHasher{}
		case "xxhash":
			h = &XXHasher{}
		default:
			return nil, fmt.Errorf("unsupported hashing algorithm: %s", cfg.HashingAlgorithm)
		}
		return NewHashBasedShortener(h, cfg.MaxCharLimit), nil
	}

	return nil, fmt.Errorf("unsupported shortener approach: %s", cfg.Approach)
}
