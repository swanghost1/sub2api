package handler

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

// ProxyNode represents a parsed proxy configuration node.
type ProxyNode struct {
	Type     string
	Name     string
	Server   string
	Port     string
	Password string
	Params   map[string]string
}

// ParseSubscriptionContent attempts to decode and parse raw subscription content.
// It supports both base64-encoded and plain-text subscription formats.
func ParseSubscriptionContent(raw string) ([]ProxyNode, error) {
	raw = strings.TrimSpace(raw)

	// Attempt base64 decode first
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		// Try URL-safe base64
		decoded, err = base64.URLEncoding.DecodeString(raw)
		if err != nil {
			// Also try RawStdEncoding (no padding) before falling back to plain text
			decoded, err = base64.RawStdEncoding.DecodeString(raw)
			if err != nil {
				// Treat as plain text
				decoded = []byte(raw)
			}
		}
	}

	lines := strings.Split(strings.TrimSpace(string(decoded)), "\n")
	var nodes []ProxyNode

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node, err := parseProxyLine(line)
		if err != nil {
			// Skip unparseable lines rather than failing the whole subscription
			continue
		}
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no valid proxy nodes found in subscription content")
	}

	return nodes, nil
}

// parseProxyLine parses a single proxy URI line into a ProxyNode.
// Supported schemes: ss://, vmess://, trojan://, vless://
func parseProxyLine(line string) (ProxyNode, error) {
	u, err := url.Parse(line)
	if err != nil {
		return ProxyNode{}, fmt.Errorf("invalid proxy URI: %w", err)
	}

	node := ProxyNode{
		Params: make(map[string]string),
	}

	switch strings.ToLower(u.Scheme) {
	case "ss":
		return parseShadowsocks(u)
	case "trojan":
		return parseTrojan(u)
	case "vmess", "vless":
		node.Type = strings.ToLower(u.Scheme)
		node.Server = u.Hostname()
		node.Port = u.Port()
		node.Name = u.Fragment
		return node, nil
	default:
		return ProxyNode{}, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
	}
}

// parseShadowsocks parses a Shadowsocks (ss://) URI.
func parseShadowsocks(u *url.URL) (ProxyNode, error) {
	node := ProxyNode{
		Type:   "ss",
		Server: u.Hostname(),
		Port:   u.Port(),
		Name:   u.Fragment,
		Params: make(map[string]string),
	}

	if u.User != nil {
		// Modern format: ss://BASE64(method:password)@host:port
		if pass, ok := u.User.Password(); ok {
			node.Params["method"] = u.User.Username()
			node.Password = pass
		} else {
			// Legacy format: ss://BASE64(method:password)@host:port
			decoded, err := base64.StdEncoding.DecodeString(u.User.Username())
			if err == nil {
				parts := strings.SplitN(string(decoded), ":", 2)
				if len(parts) == 2 {
					node.Params["method"] = parts[0]
					node.Password = parts[1]
				}
			}
		}
	}

	return node, nil
}

// parseTrojan parses a Trojan (trojan://) URI.
func parseTrojan(u *url.URL) (ProxyNode, error) {
	node := ProxyNode{
		Type:   "trojan",
		Server: u.Hostname(),
		Port:   u.Port(),
		Name:   u.Fragment,
		Params: make(map[string]string),
	}

	if u.User != nil {
		node.Password = u.User.Username()
	}

	// Parse query parameters (e.g. sni, allowInsecure)
	for key, vals := range u.Query() {
		if len(vals) > 0 {
			node.Params[key] = vals[0]
		}
	}

	return node, nil
}
