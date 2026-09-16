package shortener

import (
	"crypto/rand"

	"UrlShortner/internal/util/base62"
)

// RandomShortener uses a cryptographically secure random generator to create short codes.
type RandomShortener struct {
	charLimit int
}

func NewRandomShortener(limit int) *RandomShortener {
	return &RandomShortener{
		charLimit: limit,
	}
}

func (s *RandomShortener) Shorten(_ string) (string, error) {
	// 1. Generate enough random bytes
	// We need enough bytes to ensure base62 encoding reaches the charLimit.
	// 1 byte = max 255 (2 digits in base62). Usually limit bytes is fine.
	randomBytes := make([]byte, s.charLimit)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// 2. Base62 encode with the exact character limit
	encoded := base62.EncodeBytes(randomBytes, s.charLimit)

	return encoded, nil
}
