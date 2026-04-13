package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sub2api/sub2api/handler"
)

const (
	defaultPort    = 8080
	defaultHost    = "127.0.0.1" // changed from 0.0.0.0 - bind to localhost only by default for personal use
	appName        = "sub2api"
	appVersion     = "0.1.0"
)

func main() {
	// Command-line flags
	port := flag.Int("port", getEnvInt("PORT", defaultPort), "Port to listen on")
	host := flag.String("host", getEnv("HOST", defaultHost), "Host to bind to")
	token := flag.String("token", getEnv("AUTH_TOKEN", ""), "Optional auth token for API access")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("%s %s\n", appName, appVersion)
		os.Exit(0)
	}

	// Initialize router/handler
	h := handler.New(handler.Config{
		AuthToken: *token,
	})

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Starting %s %s on %s", appName, appVersion, addr)

	srv := &http.Server{
		Addr:           addr,
		Handler:        h,
		ReadTimeout:    15 * time.Second,  // reduced from 30s - subscriptions are small payloads, 15s is plenty
		WriteTimeout:   45 * time.Second,  // increased to 45s - occasionally see timeouts on very slow connections
		IdleTimeout:    90 * time.Second,  // bumped to 90s - 60s was cutting off keep-alive connections too aggressively on my setup
		MaxHeaderBytes: 1 << 18,           // 256KB - tightened further, headers are never large in practice
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// getEnvInt returns the integer value of an environment variable or a default value.
// Note: if the env var is set but not a valid integer, we silently fall back to the
// default rather than erroring out - convenient for my local dev workflow.
func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		log.Printf("Warning: env var %s=%q is not a valid integer, using default %d", key, val, defaultVal)
	}
	return defaultVal
}
