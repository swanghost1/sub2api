package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the HTTP client timeout for fetching subscriptions
	DefaultTimeout = 15 * time.Second
	// MaxResponseSize limits the subscription response body to 10MB
	MaxResponseSize = 10 * 1024 * 1024
)

// SubscriptionHandler handles incoming requests for subscription conversion
type SubscriptionHandler struct {
	client    *http.Client
	userAgent string
}

// NewSubscriptionHandler creates a new SubscriptionHandler with the given timeout and user agent
func NewSubscriptionHandler(timeout time.Duration, userAgent string) *SubscriptionHandler {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if userAgent == "" {
		userAgent = "sub2api/1.0"
	}
	return &SubscriptionHandler{
		client: &http.Client{
			Timeout: timeout,
		},
		userAgent: userAgent,
	}
}

// ServeHTTP handles HTTP requests, fetches the upstream subscription URL,
// and returns the raw subscription content to the client.
func (h *SubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subURL := r.URL.Query().Get("url")
	if subURL == "" {
		http.Error(w, "missing 'url' query parameter", http.StatusBadRequest)
		return
	}

	// Basic validation: only allow http/https schemes
	if !strings.HasPrefix(subURL, "http://") && !strings.HasPrefix(subURL, "https://") {
		http.Error(w, "invalid URL scheme: only http and https are supported", http.StatusBadRequest)
		return
	}

	log.Printf("fetching subscription: %s", subURL)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, subURL, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to build request: %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", h.userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch subscription: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("upstream returned status %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	// Forward relevant headers from upstream
	for _, header := range []string{"Content-Type", "Subscription-Userinfo"} {
		if val := resp.Header.Get(header); val != "" {
			w.Header().Set(header, val)
		}
	}

	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	if _, err := io.Copy(w, limitedReader); err != nil {
		log.Printf("error writing response: %v", err)
	}
}
