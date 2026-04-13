package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// buildVersion holds the version string injected at build time.
var buildVersion = "dev"

// startTime records when the process started, used for uptime calculation.
var startTime = time.Now()

// healthResponse is the JSON payload returned by the health endpoint.
type healthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	GoVersion string `json:"go_version"`
}

// HealthHandler handles HTTP requests to the health-check endpoint.
type HealthHandler struct{}

// NewHealthHandler creates and returns a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ServeHTTP writes a JSON health-check response.
// It always returns HTTP 200 as long as the process is running.
//
//	GET /health
//	{
//	  "status":     "ok",
//	  "version":    "v1.2.3",
//	  "uptime":     "3h14m15s",
//	  "go_version": "go1.22.0"
//	}
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := healthResponse{
		Status:    "ok",
		Version:   buildVersion,
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		GoVersion: runtime.Version(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	// Add X-Content-Type-Options to prevent MIME sniffing on the health response.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	// HEAD requests should not include a body.
	if r.Method == http.MethodHead {
		return
	}

	enc := json.NewEncoder(w)
	// Use compact JSON output to reduce response size.
	// Removed indentation since this endpoint is typically consumed by monitoring
	// tools rather than read manually; use `curl | jq` if pretty-printing is needed.
	if err := enc.Encode(payload); err != nil {
		// At this point the header is already sent; log only.
		_ = err
	}
}
