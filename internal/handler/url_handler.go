package handler

import (
	"encoding/json"
	"net/http"

	"UrlShortner/internal/service"
)

type URLHandler struct {
	service service.URLService
}

func NewURLHandler(s service.URLService) *URLHandler {
	return &URLHandler{service: s}
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (h *URLHandler) HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ShortenResponse{Error: "Invalid JSON payload"})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, ShortenResponse{Error: "URL is required"})
		return
	}

	shortCode, err := h.service.ShortenURL(r.Context(), req.URL)
	if err != nil {
		// Check if it's a collision or server error (simplified for now)
		writeJSON(w, http.StatusInternalServerError, ShortenResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, ShortenResponse{ShortCode: shortCode})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonBytes, _ := json.Marshal(data)
	w.Write(jsonBytes)
}
