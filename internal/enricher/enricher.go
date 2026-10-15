package enricher

import (
	"strings"
	"sync"

	"github.com/BraveRam/netpulse/internal/model"
)

// Enricher coordinates all GeoIP, ASN, flag, and reverse DNS enrichment.
type Enricher struct {
	mu          sync.Mutex
	geoIP       *GeoIPResolver
	asn         *ASNResolver
	rdns        *RDNSResolver
	pendingASNs map[string]bool
}

// NewEnricher creates a new coordinated Enricher.
func NewEnricher(geoipDBPath string) *Enricher {
	return &Enricher{
		geoIP:       NewGeoIPResolver(geoipDBPath),
		asn:         NewASNResolver(),
		rdns:        NewRDNSResolver(6),
		pendingASNs: make(map[string]bool),
	}
}

// Enrich annotates a list of connections in-place with GeoIP, ASN, flags, and rDNS.
func (e *Enricher) Enrich(conns []*model.Connection) {
	for _, c := range conns {
		e.enrichSingle(c)
	}
}

func (e *Enricher) enrichSingle(c *model.Connection) {
	// If listening socket or no remote IP
	if c.RemoteIP == nil || c.RemoteIP.IsUnspecified() || c.RemotePort == 0 || c.State == "LISTEN" {
		c.CountryCode = "LOCAL"
		c.CountryName = "Local System"
		c.FlagEmoji = "📍"
		c.ASN = "LOCAL"
		c.ASOrg = "Local System / Loopback"
		return
	}

	// 1. GeoIP resolution
	geo := e.geoIP.Resolve(c.RemoteIP)
	if geo != nil {
		c.CountryCode = geo.CountryCode
		c.CountryName = geo.CountryName
		c.City = geo.City
		c.FlagEmoji = geo.FlagEmoji
		if geo.IsLocal {
			c.ASN = "LAN"
			c.ASOrg = "Local Network / Private"
			return
		}
	}

	// 2. ASN & Organization resolution
	asInfo := e.asn.Lookup(c.RemoteIP)
	if asInfo != nil {
		c.ASN = asInfo.ASN
		c.ASOrg = asInfo.Org
		if (c.CountryCode == "" || c.CountryCode == "??") && asInfo.CountryCode != "" {
			c.CountryCode = asInfo.CountryCode
			c.CountryName = asInfo.CountryCode
			c.FlagEmoji = CountryToFlag(asInfo.CountryCode)
		}
	} else {
		// Asynchronously query Cymru ASN if IPv4 and not already queried
		e.triggerAsyncASN(c)
	}

	// Ensure a flag emoji exists
	if c.FlagEmoji == "" {
		if c.CountryCode != "" {
			c.FlagEmoji = CountryToFlag(c.CountryCode)
		} else {
			c.FlagEmoji = "🌐"
		}
	}

	// 3. Reverse DNS lookup (cached or queued)
	if c.RemoteIPStr != "" {
		domain := e.rdns.Lookup(c.RemoteIPStr)
		if domain != "" {
			c.Domain = domain
		}
	}
}

func (e *Enricher) triggerAsyncASN(c *model.Connection) {
	ipStr := c.RemoteIPStr
	if ipStr == "" || strings.Contains(ipStr, ":") {
		return // skip IPv6 for remote Cymru lookup
	}

	e.mu.Lock()
	if e.pendingASNs[ipStr] {
		e.mu.Unlock()
		return
	}
	e.pendingASNs[ipStr] = true
	e.mu.Unlock()

	ip := c.RemoteIP
	go func() {
		defer func() {
			e.mu.Lock()
			delete(e.pendingASNs, ipStr)
			e.mu.Unlock()
		}()

		_, _ = e.asn.QueryRemoteASN(ip)
	}()
}

// Close gracefully terminates background resolver workers.
func (e *Enricher) Close() {
	e.rdns.Close()
	e.geoIP.Close()
}
