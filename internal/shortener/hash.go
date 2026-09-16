package shortener

import (
	"crypto/md5"
	"crypto/sha256"
	"hash/fnv"

	"UrlShortner/internal/util/base62"

	"github.com/cespare/xxhash/v2"
	"github.com/spaolacci/murmur3"
)

// Hasher is the inner strategy interface for generating a hash from a string.
type Hasher interface {
	Hash(input string) []byte
}

// -----------------------------------------------------------------------------
// Concrete Hashers
// -----------------------------------------------------------------------------

type MD5Hasher struct{}

func (h *MD5Hasher) Hash(input string) []byte {
	sum := md5.Sum([]byte(input))
	return sum[:]
}

type SHA256Hasher struct{}

func (h *SHA256Hasher) Hash(input string) []byte {
	sum := sha256.Sum256([]byte(input))
	return sum[:]
}

type FNVHasher struct{}

func (h *FNVHasher) Hash(input string) []byte {
	hasher := fnv.New64a()
	hasher.Write([]byte(input))

	hashBucket := make([]byte, 0, 8)

	return hasher.Sum(hashBucket)
}

type MurmurHasher struct{}

func (h *MurmurHasher) Hash(input string) []byte {
	hasher := murmur3.New64()
	hasher.Write([]byte(input))
	hashBucket := make([]byte, 0, 8)

	return hasher.Sum(hashBucket)
}

type XXHasher struct{}

func (h *XXHasher) Hash(input string) []byte {
	hasher := xxhash.New()
	hasher.Write([]byte(input))
	hashBucket := make([]byte, 0, 8)

	return hasher.Sum(hashBucket)
}

// -----------------------------------------------------------------------------
// Hash-Based Shortener Strategy
// -----------------------------------------------------------------------------

// HashBasedShortener uses a specific Hasher to hash the URL, then encodes it with Base62.
type HashBasedShortener struct {
	hasher    Hasher
	charLimit int
}

func NewHashBasedShortener(hasher Hasher, limit int) *HashBasedShortener {
	return &HashBasedShortener{
		hasher:    hasher,
		charLimit: limit,
	}
}

// Shorten executes the hash-based strategy.
func (s *HashBasedShortener) Shorten(longURL string) (string, error) {
	// 1. Hash the long URL
	hashBytes := s.hasher.Hash(longURL)

	// 2. Base62 encode the resulting bytes with the specific character limit
	encoded := base62.EncodeBytes(hashBytes, s.charLimit)

	return encoded, nil
}
