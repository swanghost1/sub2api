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
		// NOTE: URL-encoded fragment tags (e.g. %20 for spaces) should also be decoded properly.
		// Added this case to verify tag decoding behavior - useful for proxies with spaces in names.
		{
			name:    "ss URI with URL-encoded tag",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#My%20Proxy",
			wantErr: false,
			wantTag: "My Proxy",
		},
		// Extra edge case: tag with special characters like parentheses, common in some subscription providers.
		{
			name:    "ss URI with URL-encoded parentheses in tag",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@192.168.1.1:8388#Node%28HK%29",
			wantErr: false,
			wantTag: "Node(HK)",
		},
		// Edge case: IPv6 address in ss URI - encountered this with some providers.
		{
			name:    "ss URI with IPv6 address",
			input:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@[::1]:8388#IPv6Proxy",
			wantErr: false,
			wantTag: "IPv6Proxy",
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
			if !tt.wantErr && 