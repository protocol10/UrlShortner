package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockShortener struct {
	MockShortCode string
	MockError     error
}

// Implement the interface method
func (m *MockShortener) Shorten(longURL string) (string, error) {
	return m.MockShortCode, m.MockError
}

type MockURLRepository struct {
	MockError   error
	MockLongURL string
}

// Implement the interface method
func (m *MockURLRepository) Insert(ctx context.Context, longURL, shortCode string) error {
	return m.MockError // Just return whatever we tell the mock to return in the test!
}

func (m *MockURLRepository) GetByShortCode(ctx context.Context, shortCode string) (string, error) {
	return m.MockLongURL, m.MockError
}

func Test_urlService_ShortenURL(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		longURL string
		want    string
		wantErr bool
	}{
		{
			name:    "Success case",
			longURL: "https://www.example.com/some/long/path",
			want:    "exmpl123",
			wantErr: false,
		},
		{
			name:    "Collision on first try",
			longURL: "https://occupied.com",
			want:    "collision",
			wantErr: false, // Should retry once and succeed
		},
		{
			name:    "Permanent Collision",
			longURL: "https://permanentcollision.com",
			want:    "",   // Returns empty string on failure
			wantErr: true, // Returns error because it failed all retries
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			mockRepo := &MockURLRepository{}
			var expectedErr error
			if tt.wantErr {
				expectedErr = errors.New("Code already exist")
			}
			// We set the mock shortener to return what the test case wants
			mockShortener := &MockShortener{
				MockShortCode: tt.want,
				MockError:     expectedErr,
			}

			// Create the real URLService, injecting our mocks
			svc := NewURLService(mockShortener, mockRepo)

			got, gotErr := svc.ShortenURL(context.Background(), tt.longURL)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ShortenURL() failed: %v", gotErr)
				}
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, gotErr != nil)
		})
	}
}

func Test_urlService_GetLongURL(t *testing.T) {
	tests := []struct {
		name        string
		shortCode   string
		mockLongURL string
		mockErr     error
		want        string
		wantErr     bool
	}{
		{
			name:        "Success",
			shortCode:   "googl1",
			mockLongURL: "https://www.google.com",
			mockErr:     nil,
			want:        "https://www.google.com",
			wantErr:     false,
		},
		{
			name:        "Empty short code",
			shortCode:   "",
			mockLongURL: "",
			mockErr:     nil,
			want:        "",
			wantErr:     true,
		},
		{
			name:        "Repo error",
			shortCode:   "notfound",
			mockLongURL: "",
			mockErr:     errors.New("not found"),
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockURLRepository{
				MockLongURL: tt.mockLongURL,
				MockError:   tt.mockErr,
			}
			mockShortener := &MockShortener{}

			svc := NewURLService(mockShortener, mockRepo)

			got, err := svc.GetLongURL(context.Background(), tt.shortCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
