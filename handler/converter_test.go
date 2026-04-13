package handler

import (
	"testing"
)

// TestParseShadowsocks tests the Shadowsocks URI parser.
func TestParseShadowsocks(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantTag string
	}{
		{
			name:    "valid base64 encoded ss URI",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#MyProxy",
			wantErr: false,
			wantTag: "MyProxy",
		},
		{
			name:    "valid SIP002 ss URI",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388",
			wantErr: false,
			wantTag: "",
		},
		{
			name:    "invalid ss URI missing prefix",
			input:   "YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseShadowsocks(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseShadowsocks(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && proxy == nil {
				t.Errorf("parseShadowsocks(%q) returned nil proxy without error", tt.input)
			}
			if !tt.wantErr && tt.wantTag != "" && proxy["tag"] != tt.wantTag {
				t.Errorf("parseShadowsocks(%q) tag = %v, want %v", tt.input, proxy["tag"], tt.wantTag)
			}
		})
	}
}

// TestParseTrojan tests the Trojan URI parser.
func TestParseTrojan(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantTag string
	}{
		{
			name:    "valid trojan URI with tag",
			input:   "trojan://password@example.com:443?sni=example.com#MyTrojan",
			wantErr: false,
			wantTag: "MyTrojan",
		},
		{
			name:    "valid trojan URI without tag",
			input:   "trojan://password@example.com:443",
			wantErr: false,
			wantTag: "",
		},
		{
			name:    "invalid trojan URI missing prefix",
			input:   "password@example.com:443",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseTrojan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTrojan(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && proxy == nil {
				t.Errorf("parseTrojan(%q) returned nil proxy without error", tt.input)
			}
			if !tt.wantErr && tt.wantTag != "" && proxy["tag"] != tt.wantTag {
				t.Errorf("parseTrojan(%q) tag = %v, want %v", tt.input, proxy["tag"], tt.wantTag)
			}
		})
	}
}

// TestParseSubscriptionContent tests the top-level subscription content parser.
func TestParseSubscriptionContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantCount  int
		wantErr    bool
	}{
		{
			name:      "empty content",
			content:   "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "content with only comments",
			content:   "# this is a comment\n// another comment",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "mixed valid and invalid lines",
			content:   "trojan://password@example.com:443#Test\ninvalidline\n",
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxies, err := ParseSubscriptionContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSubscriptionContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(proxies) != tt.wantCount {
				t.Errorf("ParseSubscriptionContent() returned %d proxies, want %d", len(proxies), tt.wantCount)
			}
		})
	}
}
