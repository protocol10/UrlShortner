package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"UrlShortner/internal/repository"

	"github.com/stretchr/testify/assert"
)

// 1. Create a Mock Service that implements service.URLService
type MockURLService struct {
	MockShortCode string
	MockLongURL   string
	MockError     error
}

func (m *MockURLService) ShortenURL(ctx context.Context, longURL string) (string, error) {
	return m.MockShortCode, m.MockError
}

func (m *MockURLService) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	return m.MockLongURL, m.MockError
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

func TestHandleGet_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/googl1", nil)
	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{}
	h := NewURLHandler(mockSvc)

	h.HandleGet(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleGet_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten?short_code=googl1", nil)
	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{
		MockLongURL: "https://www.google.com",
		MockError:   nil,
	}

	h := NewURLHandler(mockSvc)
	h.HandleGet(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "https://www.google.com", rr.Header().Get("Location"))
}

func TestHandleGet_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten?short_code=unknown", nil)
	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{
		MockError: repository.ErrNotFound,
	}

	h := NewURLHandler(mockSvc)
	h.HandleGet(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	expectedBody := `{"error":"Short code not found"}`
	assert.Equal(t, expectedBody, rr.Body.String())
}

func TestHandleGet_MissingShortCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shorten", nil)
	rr := httptest.NewRecorder()

	mockSvc := &MockURLService{}
	h := NewURLHandler(mockSvc)

	h.HandleGet(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	expectedBody := `{"error":"Short code is required"}`
	assert.Equal(t, expectedBody, rr.Body.String())
}
