package enricher

import (
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/BraveRam/netpulse/internal/model"
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

func TestGeoIPDatabase(t *testing.T) {
	for _, city := range []bool{false, true} {
		name := "country"
		if city {
			name = "city"
		}
		t.Run(name, func(t *testing.T) {
			tree, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: "GeoIP2-Test", IPVersion: 6})
			if err != nil {
				t.Fatal(err)
			}
			record := mmdbtype.Map{"country": mmdbtype.Map{"iso_code": mmdbtype.String("ET"), "names": mmdbtype.Map{"en": mmdbtype.String("Ethiopia")}}}
			if city {
				record["city"] = mmdbtype.Map{"names": mmdbtype.Map{"en": mmdbtype.String("Addis Ababa")}}
			}
			for _, cidr := range []string{"8.8.8.0/24", "2001:4860::/32"} {
				_, network, err := net.ParseCIDR(cidr)
				if err != nil {
					t.Fatal(err)
				}
				if err := tree.Insert(network, record); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(t.TempDir(), "test.mmdb")
			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tree.WriteTo(file); err != nil {
				file.Close()
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			resolver := NewGeoIPResolver(path)
			defer resolver.Close()
			for _, ip := range []string{"8.8.8.8", "2001:4860::1"} {
				info := resolver.Resolve(net.ParseIP(ip))
				if info.CountryCode != "ET" || info.CountryName != "Ethiopia" || info.FlagEmoji != "🇪🇹" {
					t.Fatalf("database record not resolved: %+v", info)
				}
				if city && info.City != "Addis Ababa" {
					t.Fatalf("city not resolved: %+v", info)
				}
			}
			if info := resolver.Resolve(net.ParseIP("9.9.9.9")); info.CountryCode != "??" {
				t.Fatalf("unexpected missing-IP result: %+v", info)
			}
			if info := resolver.Resolve(net.ParseIP("127.0.0.1")); !info.IsLocal {
				t.Fatal("local IP no longer classified locally")
			}
		})
	}
}

func TestGeoIPMissingDatabaseFallsBack(t *testing.T) {
	resolver := NewGeoIPResolver(filepath.Join(t.TempDir(), "missing.mmdb"))
	defer resolver.Close()
	if info := resolver.Resolve(net.ParseIP("8.8.8.8")); info.CountryCode != "??" {
		t.Fatalf("unexpected fallback: %+v", info)
	}
}
