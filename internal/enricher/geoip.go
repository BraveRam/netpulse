package enricher

import (
	"net"
	"os"
)

// GeoIPInfo holds geographic metadata for an IP.
type GeoIPInfo struct {
	CountryCode string
	CountryName string
	City        string
	FlagEmoji   string
	IsLocal     bool
}

// GeoIPResolver resolves geographic location for IP addresses.
type GeoIPResolver struct {
	dbPath string
}

// NewGeoIPResolver creates a GeoIPResolver.
func NewGeoIPResolver(customDBPath string) *GeoIPResolver {
	dbPath := customDBPath
	if dbPath == "" {
		// Auto-discover common locations
		candidates := []string{
			"/var/lib/GeoIP/GeoLite2-City.mmdb",
			"/var/lib/GeoIP/GeoLite2-Country.mmdb",
			"/usr/share/GeoIP/GeoLite2-City.mmdb",
			"/usr/share/GeoIP/GeoLite2-Country.mmdb",
		}
		for _, cand := range candidates {
			if _, err := os.Stat(cand); err == nil {
				dbPath = cand
				break
			}
		}
	}

	return &GeoIPResolver{
		dbPath: dbPath,
	}
}

// Resolve looks up country, city, and flag for an IP.
func (g *GeoIPResolver) Resolve(ip net.IP) *GeoIPInfo {
	if ip == nil {
		return &GeoIPInfo{
			CountryCode: "??",
			CountryName: "Unknown",
			FlagEmoji:   "🌐",
		}
	}

	// 1. Check for local / private / loopback addresses
	if isLocalOrPrivate(ip) {
		return &GeoIPInfo{
			CountryCode: "LOCAL",
			CountryName: "Local Network",
			City:        "LAN",
			FlagEmoji:   CountryToFlag("LOCAL"),
			IsLocal:     true,
		}
	}

	// Default fallback for public IPs without database
	return &GeoIPInfo{
		CountryCode: "??",
		CountryName: "Global",
		FlagEmoji:   "🌐",
		IsLocal:     false,
	}
}

// isLocalOrPrivate checks whether an IP is loopback, RFC1918 private, link-local, or multicast.
func isLocalOrPrivate(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		// 10.0.0.0/8
		if ip4[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// 127.0.0.0/8
		if ip4[0] == 127 {
			return true
		}
		// 169.254.0.0/16 (Link Local)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		// 100.64.0.0/10 (Carrier-Grade NAT)
		if ip4[0] == 100 && (ip4[1]&0xC0) == 64 {
			return true
		}
	} else {
		// IPv6 Unique Local Address (fc00::/7)
		if len(ip) == 16 && (ip[0]&0xfe) == 0xfc {
			return true
		}
	}

	return false
}
