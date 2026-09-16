package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 1. Create a Mock Service that implements service.URLService
type MockURLService struct {
	MockShortCode string
	MockError     error
}

func (m *MockURLService) ShortenURL(ctx context.Context, longURL string) (string, error) {
	return m.MockShortCode, m.MockError
}

// 2. Our first test function
func TestHandleShorten_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten", nil)

	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{}
	h := NewURLHandler(mockSvc)

	h.HandleShorten(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d (Method Not Allowed), got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandlerShorten_EmptyUrl(t *testing.T) {

	var requestBody = `{
    "url": ""
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(requestBody))

	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{}

	h := NewURLHandler(mockSvc)
	h.HandleShorten(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	expectedBody := `{"error":"URL is required"}`
	assert.Equal(t, expectedBody, rr.Body.String())
}

func TestHandlerShorten_Success(t *testing.T) {
	var requestBody = `{
    "url": "https://www.example.com/some/long/path"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(requestBody))

	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{
		MockShortCode: "exmpl123",
		MockError:     nil, // No error means success!
	}

	h := NewURLHandler(mockSvc)
	h.HandleShorten(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// 6. Optional: Assert the response body contains the expected short code
	expectedBody := `{"short_code":"exmpl123"}`
	assert.Equal(t, expectedBody, rr.Body.String())
}

func TestHandlerShorten_InternalServerError(t *testing.T) {
	var requestBody = `{
    "url": "https://www.example.com/some/long/path"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(requestBody))

	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{
		MockShortCode: "",
		MockError:     errors.New("internal error"), // No error means success!
	}

	h := NewURLHandler(mockSvc)
	h.HandleShorten(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// 6. Optional: Assert the response body contains the expected short code
	expectedBody := `{"error":"internal error"}`
	assert.Equal(t, expectedBody, rr.Body.String())
}
