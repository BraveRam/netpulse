package enricher

import (
	"net"
	"testing"

	"github.com/plxor/netpulse/internal/model"
)

func TestCountryToFlag(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"US", "🇺🇸"},
		{"DE", "🇩🇪"},
		{"JP", "🇯🇵"},
		{"GB", "🇬🇧"},
		{"LOCAL", "📍"},
		{"", "🌐"},
		{"??", "🌐"},
	}

	for _, tt := range tests {
		got := CountryToFlag(tt.code)
		if got != tt.expected {
			t.Errorf("CountryToFlag(%q) = %q; expected %q", tt.code, got, tt.expected)
		}
	}
}

func TestGeoIPResolverLocal(t *testing.T) {
	resolver := NewGeoIPResolver("")

	localIPs := []string{
		"127.0.0.1",
		"10.0.0.5",
		"192.168.1.100",
		"172.16.0.1",
		"::1",
	}

	for _, ipStr := range localIPs {
		ip := net.ParseIP(ipStr)
		info := resolver.Resolve(ip)
		if !info.IsLocal {
			t.Errorf("Expected IP %s to be resolved as local", ipStr)
		}
		if info.FlagEmoji != "📍" {
			t.Errorf("Expected flag '📍' for IP %s, got %s", ipStr, info.FlagEmoji)
		}
	}
}

func TestASNResolverStatic(t *testing.T) {
	resolver := NewASNResolver()

	tests := []struct {
		ipStr       string
		expectedASN string
		expectedOrg string
	}{
		{"1.1.1.1", "AS13335", "Cloudflare, Inc."},
		{"8.8.8.8", "AS15169", "Google LLC"},
		{"17.253.144.10", "AS714", "Apple Inc."},
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ipStr)
		info := resolver.Lookup(ip)
		if info == nil {
			t.Fatalf("Lookup(%s) returned nil", tt.ipStr)
		}
		if info.ASN != tt.expectedASN {
			t.Errorf("For %s, expected ASN %s, got %s", tt.ipStr, tt.expectedASN, info.ASN)
		}
		if info.Org != tt.expectedOrg {
			t.Errorf("For %s, expected Org %s, got %s", tt.ipStr, tt.expectedOrg, info.Org)
		}
	}
}

func TestEnricher(t *testing.T) {
	enricher := NewEnricher("")
	defer enricher.Close()

	conn := &model.Connection{
		RemoteIP:    net.ParseIP("1.1.1.1"),
		RemoteIPStr: "1.1.1.1",
		RemotePort:  443,
		State:       "ESTABLISHED",
	}

	enricher.Enrich([]*model.Connection{conn})

	if conn.ASN != "AS13335" {
		t.Errorf("Expected ASN AS13335, got %s", conn.ASN)
	}
	if conn.FlagEmoji == "" {
		t.Errorf("Expected non-empty flag emoji")
	}
}
