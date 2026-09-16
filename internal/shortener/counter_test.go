package shortener

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockCounterClient struct {
	MockID  int64
	MockErr error
}

func (m *MockCounterClient) Incr(_ context.Context, _ string) (int64, error) {
	return m.MockID, m.MockErr
}

func TestCounterShortener_Shorten(t *testing.T) {
	defaultCfg := CounterShortenerConfig{
		CounterKey:   "test_key",
		Multiplier:   11400714819323198485,
		XORMask:      0x5BF03635467C061D,
		MaxCharLimit: 6,
	}

	tests := []struct {
		name         string
		mockClient   CounterClient
		cfg          CounterShortenerConfig
		longURL      string
		wantLen      int
		wantErr      bool
		errSubstring string
	}{
		{
			name: "Success - Counter ID 1",
			mockClient: &MockCounterClient{
				MockID:  1,
				MockErr: nil,
			},
			cfg:          defaultCfg,
			longURL:      "https://google.com",
			wantLen:      6,
			wantErr:      false,
			errSubstring: "",
		},
		{
			name: "Success - Counter ID 2",
			mockClient: &MockCounterClient{
				MockID:  2,
				MockErr: nil,
			},
			cfg:          defaultCfg,
			longURL:      "https://facebook.com",
			wantLen:      6,
			wantErr:      false,
			errSubstring: "",
		},
		{
			name: "Failure - Redis Connection Error",
			mockClient: &MockCounterClient{
				MockID:  0,
				MockErr: errors.New("redis connection refused"),
			},
			cfg:          defaultCfg,
			longURL:      "https://google.com",
			wantLen:      0,
			wantErr:      true,
			errSubstring: "redis connection refused",
		},
		{
			name:         "Failure - Uninitialized Client",
			mockClient:   nil,
			cfg:          defaultCfg,
			longURL:      "https://google.com",
			wantLen:      0,
			wantErr:      true,
			errSubstring: "redis counter client is not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := NewCounterShortener(tt.mockClient, tt.cfg)
			code, err := shortener.Shorten(tt.longURL)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, code)
				if tt.errSubstring != "" {
					assert.Contains(t, err.Error(), tt.errSubstring)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, code)
				assert.Equal(t, tt.wantLen, len(code))
			}
		})
	}
}

func TestCounterShortener_DistinctOutputs(t *testing.T) {
	mockClient := &MockCounterClient{MockID: 1}
	cfg := CounterShortenerConfig{
		CounterKey:   "test_key",
		Multiplier:   11400714819323198485,
		XORMask:      0x5BF03635467C061D,
		MaxCharLimit: 6,
	}

	shortener := NewCounterShortener(mockClient, cfg)

	code1, err1 := shortener.Shorten("https://google.com")
	assert.NoError(t, err1)

	mockClient.MockID = 2
	code2, err2 := shortener.Shorten("https://google.com")
	assert.NoError(t, err2)

	assert.NotEqual(t, code1, code2, "Sequential IDs must produce distinct short codes")
}

func BenchmarkCounterShortener_Shorten(b *testing.B) {
	mockClient := &MockCounterClient{MockID: 1000000}
	cfg := CounterShortenerConfig{
		CounterKey:   "benchmark_key",
		Multiplier:   11400714819323198485,
		XORMask:      0x5BF03635467C061D,
		MaxCharLimit: 6,
	}

	shortener := NewCounterShortener(mockClient, cfg)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = shortener.Shorten("https://www.example.com")
	}
}
