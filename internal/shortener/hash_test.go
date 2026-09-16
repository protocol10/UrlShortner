package shortener

import (
	"testing"
)

// A standard length URL to test against
var longURL = "https://www.example.com/some/very/long/path/that/needs/to/be/shortened?id=123456789&user=test"

func BenchmarkMD5(b *testing.B) {
	hasher := &MD5Hasher{}
	for i := 0; i < b.N; i++ {
		hasher.Hash(longURL)
	}
}

func BenchmarkSHA256(b *testing.B) {
	hasher := &SHA256Hasher{}
	for i := 0; i < b.N; i++ {
		hasher.Hash(longURL)
	}
}

func BenchmarkFNV(b *testing.B) {
	hasher := &FNVHasher{}
	for i := 0; i < b.N; i++ {
		hasher.Hash(longURL)
	}
}

func BenchmarkMurmur(b *testing.B) {
	hasher := &MurmurHasher{}
	for i := 0; i < b.N; i++ {
		hasher.Hash(longURL)
	}
}

func BenchmarkXXHash(b *testing.B) {
	hasher := &XXHasher{}
	for i := 0; i < b.N; i++ {
		hasher.Hash(longURL)
	}
}
