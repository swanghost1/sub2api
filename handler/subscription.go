package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscriptionHandler handles subscription-related HTTP requests.
type SubscriptionHandler struct {
	client *http.Client
}

// NewSubscriptionHandler creates a new SubscriptionHandler with a default HTTP client.
func NewSubscriptionHandler() *SubscriptionHandler {
	return &SubscriptionHandler{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetSubscription fetches a remote subscription URL and returns its content
// as a JSON-compatible response, parsing each proxy line into a structured object.
//
// Query parameters:
//   - url: the remote subscription URL to fetch (required)
//   - target: output format hint, e.g. "clash", "v2ray" (optional, default: raw)
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	subURL := c.Query("url")
	if subURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing required query parameter: url",
		})
		return
	}

	target := strings.ToLower(c.DefaultQuery("target", "raw"))

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, subURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid subscription URL: %v", err),
		})
		return
	}

	// Mimic a common subscription client user-agent so remote servers respond correctly.
	req.Header.Set("User-Agent", "clash/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": fmt.Sprintf("failed to fetch subscription: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": fmt.Sprintf("upstream returned status %d", resp.StatusCode),
		})
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to read response body: %v", err),
		})
		return
	}

	switch target {
	case "raw":
		// Return the raw subscription content with appropriate content type.
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/plain; charset=utf-8"
		}
		c.Data(http.StatusOK, contentType, body)
	default:
		// For unsupported targets, return raw content and note the limitation.
		c.JSON(http.StatusOK, gin.H{
			"target":  target,
			"warning": fmt.Sprintf("target format %q is not yet supported; returning raw content", target),
			"content": string(body),
		})
	}
}
