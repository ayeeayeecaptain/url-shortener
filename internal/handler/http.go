package handler

import (
	"context" // <--- Make sure this is imported
	"encoding/json"
	"net/http"
	"strings"
)

type Service interface {
	Shorten(ctx context.Context, longURL string) (string, error) // <--- Fixed here
	Resolve(ctx context.Context, token string) (string, error)   // <--- Fixed here
}

type HTTPHandler struct {
	service Service
}

func NewHTTPHandler(s Service) *HTTPHandler {
	return &HTTPHandler{service: s}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost && r.URL.Path == "/shorten" {
		h.handleShorten(w, r)
		return
	} else if r.Method == http.MethodGet {
		h.handleResolve(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
}

type shortenRequest struct {
	URL string `json:"url"`
}

func (h *HTTPHandler) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid payload payload request"})
		return
	}

	token, err := h.service.Shorten(r.Context(), req.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"short_url": "/" + token})
}

func (h *HTTPHandler) handleResolve(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/")
	if token == "" || len(token) > 10 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid short token"})
		return
	}

	longURL, err := h.service.Resolve(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "url entry not found"})
		return
	}

	// 301 Permanent Redirect represents correct production handling semantic for shorteners
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}
