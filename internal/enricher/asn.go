package enricher

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// ASInfo represents Autonomous System metadata and organization.
type ASInfo struct {
	ASN         string
	Org         string
	CountryCode string
}

// ASNResolver resolves Autonomous System metadata using built-in CIDRs,
// LRU cache, and DNS TXT lookups (Team Cymru origin.asn.cymru.com).
type ASNResolver struct {
	mu          sync.RWMutex
	cache       map[string]*ASInfo
	staticCIDRs []cidrEntry
}

type cidrEntry struct {
	ipNet *net.IPNet
	info  ASInfo
}

// NewASNResolver creates an ASNResolver with built-in prefixes.
func NewASNResolver() *ASNResolver {
	r := &ASNResolver{
		cache: make(map[string]*ASInfo),
	}
	r.loadStaticCIDRs()
	return r
}

func (r *ASNResolver) loadStaticCIDRs() {
	staticList := []struct {
		cidr string
		asn  string
		org  string
		cc   string
	}{
		// Cloudflare
		{"1.1.1.0/24", "AS13335", "Cloudflare, Inc.", "US"},
		{"1.0.0.0/24", "AS13335", "Cloudflare, Inc.", "US"},
		{"104.16.0.0/12", "AS13335", "Cloudflare, Inc.", "US"},
		{"172.64.0.0/13", "AS13335", "Cloudflare, Inc.", "US"},
		{"162.158.0.0/15", "AS13335", "Cloudflare, Inc.", "US"},
		{"198.41.128.0/17", "AS13335", "Cloudflare, Inc.", "US"},
		// Google
		{"8.8.8.0/24", "AS15169", "Google LLC", "US"},
		{"8.8.4.0/24", "AS15169", "Google LLC", "US"},
		{"142.250.0.0/15", "AS15169", "Google LLC", "US"},
		{"172.217.0.0/16", "AS15169", "Google LLC", "US"},
		{"216.58.192.0/19", "AS15169", "Google LLC", "US"},
		{"74.125.0.0/16", "AS15169", "Google LLC", "US"},
		{"173.194.0.0/16", "AS15169", "Google LLC", "US"},
		// Amazon AWS
		{"3.0.0.0/9", "AS16509", "Amazon.com, Inc.", "US"},
		{"52.0.0.0/11", "AS16509", "Amazon.com, Inc.", "US"},
		{"54.0.0.0/12", "AS16509", "Amazon.com, Inc.", "US"},
		{"13.32.0.0/12", "AS16509", "Amazon.com, Inc.", "US"},
		{"99.84.0.0/16", "AS16509", "Amazon.com, Inc.", "US"},
		// Microsoft / Azure
		{"20.0.0.0/11", "AS8075", "Microsoft Corporation", "US"},
		{"40.74.0.0/15", "AS8075", "Microsoft Corporation", "US"},
		{"51.140.0.0/14", "AS8075", "Microsoft Corporation", "US"},
		{"13.64.0.0/11", "AS8075", "Microsoft Corporation", "US"},
		// Apple
		{"17.0.0.0/8", "AS714", "Apple Inc.", "US"},
		// GitHub
		{"140.82.112.0/20", "AS36459", "GitHub, Inc.", "US"},
		{"192.30.252.0/22", "AS36459", "GitHub, Inc.", "US"},
		{"185.199.108.0/22", "AS36459", "GitHub, Inc.", "US"},
		// Fastly
		{"151.101.0.0/16", "AS54113", "Fastly, Inc.", "US"},
		{"199.232.0.0/16", "AS54113", "Fastly, Inc.", "US"},
		// Akamai
		{"23.0.0.0/12", "AS16625", "Akamai Technologies", "US"},
		{"104.64.0.0/10", "AS16625", "Akamai Technologies", "US"},
		// Quad9
		{"9.9.9.0/24", "AS19281", "Quad9", "CH"},
		// Cisco OpenDNS
		{"208.67.220.0/22", "AS36692", "Cisco OpenDNS", "US"},
		// Telegram
		{"149.154.160.0/20", "AS62041", "Telegram Messenger Inc", "GB"},
		{"91.108.4.0/22", "AS62041", "Telegram Messenger Inc", "GB"},
		// Meta / Facebook
		{"157.240.0.0/16", "AS32934", "Meta Platforms, Inc.", "US"},
		{"31.13.64.0/18", "AS32934", "Meta Platforms, Inc.", "US"},
		// DigitalOcean
		{"104.131.0.0/16", "AS14061", "DigitalOcean, LLC", "US"},
		{"159.203.0.0/16", "AS14061", "DigitalOcean, LLC", "US"},
		{"167.99.0.0/16", "AS14061", "DigitalOcean, LLC", "US"},
		// Hetzner
		{"78.46.0.0/15", "AS24940", "Hetzner Online GmbH", "DE"},
		{"88.198.0.0/16", "AS24940", "Hetzner Online GmbH", "DE"},
		{"168.119.0.0/16", "AS24940", "Hetzner Online GmbH", "DE"},
		// OVH
		{"51.254.0.0/15", "AS16276", "OVH SAS", "FR"},
		{"147.135.0.0/16", "AS16276", "OVH SAS", "FR"},
	}

	for _, item := range staticList {
		_, ipNet, err := net.ParseCIDR(item.cidr)
		if err == nil {
			r.staticCIDRs = append(r.staticCIDRs, cidrEntry{
				ipNet: ipNet,
				info: ASInfo{
					ASN:         item.asn,
					Org:         item.org,
					CountryCode: item.cc,
				},
			})
		}
	}
}

// Lookup attempts to resolve the ASInfo for an IP.
// Returns cached/static info if available, or nil if unknown.
func (r *ASNResolver) Lookup(ip net.IP) *ASInfo {
	if ip == nil {
		return nil
	}

	ipStr := ip.String()

	r.mu.RLock()
	if info, ok := r.cache[ipStr]; ok {
		r.mu.RUnlock()
		return info
	}
	r.mu.RUnlock()

	// Check static CIDRs
	for _, entry := range r.staticCIDRs {
		if entry.ipNet.Contains(ip) {
			info := &entry.info
			r.mu.Lock()
			r.cache[ipStr] = info
			r.mu.Unlock()
			return info
		}
	}

	return nil
}

// QueryRemoteASN queries Team Cymru DNS for the ASN and Organization if not cached.
// Should be invoked in background goroutine.
func (r *ASNResolver) QueryRemoteASN(ip net.IP) (*ASInfo, error) {
	if ip == nil {
		return nil, fmt.Errorf("nil IP")
	}

	ip4 := ip.To4()
	if ip4 == nil {
		// IPv6 Cymru lookup format is complex nibble reversal; skip remote for now
		return nil, fmt.Errorf("IPv6 remote query unsupported")
	}

	ipStr := ip4.String()
	// Reversal: a.b.c.d -> d.c.b.a.origin.asn.cymru.com
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid IPv4 string")
	}

	reverseDomain := fmt.Sprintf("%s.%s.%s.%s.origin.asn.cymru.com", parts[3], parts[2], parts[1], parts[0])

	records, err := net.LookupTXT(reverseDomain)
	if err != nil || len(records) == 0 {
		return nil, err
	}

	// Format: "15169 | 8.8.8.0/24 | US | arin | 1992-12-01"
	line := records[0]
	segments := strings.Split(line, "|")
	if len(segments) < 3 {
		return nil, fmt.Errorf("unexpected Cymru response format: %s", line)
	}

	asnNumber := strings.TrimSpace(segments[0])
	countryCode := strings.TrimSpace(segments[2])
	asnFormatted := "AS" + asnNumber

	// Optional: query AS name from AS<number>.asn.cymru.com
	orgName := asnFormatted
	asRecords, err := net.LookupTXT(fmt.Sprintf("%s.asn.cymru.com", asnFormatted))
	if err == nil && len(asRecords) > 0 {
		// Format: "15169 | US | arin | 1992-12-01 | GOOGLE, US"
		asSegments := strings.Split(asRecords[0], "|")
		if len(asSegments) >= 5 {
			orgName = strings.TrimSpace(asSegments[4])
		}
	}

	info := &ASInfo{
		ASN:         asnFormatted,
		Org:         orgName,
		CountryCode: countryCode,
	}

	r.mu.Lock()
	r.cache[ipStr] = info
	r.mu.Unlock()

	return info, nil
}

// Put caches ASInfo for an IP.
func (r *ASNResolver) Put(ipStr string, info *ASInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.cache) > 10000 {
		// Simple cache eviction: clear half
		for k := range r.cache {
			delete(r.cache, k)
			if len(r.cache) <= 5000 {
				break
			}
		}
	}
	r.cache[ipStr] = info
}
