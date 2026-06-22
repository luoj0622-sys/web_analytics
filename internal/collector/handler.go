package collector

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Collect(w http.ResponseWriter, r *http.Request) {
	var req CollectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Client = ClientMetadata{
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	}

	if _, err := h.service.Collect(r.Context(), req); err != nil {
		if errors.Is(err, ErrInvalidCredential) {
			http.Error(w, "invalid credential", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		host, _, err := net.SplitHostPort(forwarded)
		if err == nil {
			return host
		}
		return forwarded
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
