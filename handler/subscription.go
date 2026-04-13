package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscriptionHandler handles subscription conversion requests
type SubscriptionHandler struct {
	client *http.Client
}

// NewSubscriptionHandler creates a new SubscriptionHandler with a configured HTTP client
func NewSubscriptionHandler(timeoutSeconds int) *SubscriptionHandler {
	return &SubscriptionHandler{
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

// Convert handles GET /convert?url=<subscription_url>&target=<target_format>
// It fetches the remote subscription and returns the raw content to the client.
func (h *SubscriptionHandler) Convert(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameter: url"})
		return
	}

	// Validate the provided URL
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url: must be http or https"})
		return
	}

	// Default target changed to "sing-box" since that's what I primarily use
	target := strings.ToLower(c.DefaultQuery("target", "sing-box"))

	// Build upstream request
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to build request: %v", err)})
		return
	}

	// Forward a realistic User-Agent so upstream servers don't block the request
	req.Header.Set("User-Agent", subscriptionUserAgent(target))

	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to fetch subscription: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":       "upstream returned non-200 status",
			"status_code": resp.StatusCode,
		})
		return
	}

	// Increased limit to 20 MB — some of my subscriptions with many nodes were getting truncated
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to read response body: %v", err)})
		return
	}

	// Propagate relevant upstream headers
	for _, header := range []string{
		"Content-Disposition",
		"Subscription-Userinfo",
		"Profile-Update-Interval",
		"Profile-Title",
	} {
		if val := resp.Header.Get(header); val != "" {
			c.Header(header, val)
		}
	}

	contentType := resolveContentType(target)
	c.Data(http.StatusOK, contentType, body)
}

// subscriptionUserAgent returns an appropriate User-Agent string for the given target format.
func subscriptionUserAgent(target string) string {
	switch target {
	case "clash", "clashr":
		return "ClashforWindows/0.20.39"
	case "surge":
		return "Surge/2023"
	case "quantumult", "quantumultx":
		return "Quantumult/1.0"
	// sing-box: use a recent version string; update this when the client updates
	case "sing-box":
		return "sing-box/1.9.0"
	default:
		return "sub2api/1.0"
	}
}
