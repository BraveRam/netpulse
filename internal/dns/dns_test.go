package dns

import (
	"testing"
	"time"

	"github.com/plxor/netpulse/internal/model"
)

func TestParseDNSPacketQuery(t *testing.T) {
	// Sample DNS query packet for "example.com" Type A
	// ID: 0x1234, Flags: 0x0100 (standard query, RD), QDCOUNT: 1, ANCOUNT: 0, NSCOUNT: 0, ARCOUNT: 0
	// Question: 7 "example" 3 "com" 0, Type: 0x0001 (A), Class: 0x0001 (IN)
	packet := []byte{
		0x12, 0x34, // ID
		0x01, 0x00, // Flags (Query)
		0x00, 0x01, // QDCOUNT = 1
		0x00, 0x00, // ANCOUNT = 0
		0x00, 0x00, // NSCOUNT = 0
		0x00, 0x00, // ARCOUNT = 0
		// QNAME: 7example3com0
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,
		0x00, 0x01, // QTYPE = A
		0x00, 0x01, // QCLASS = IN
	}

	q, err := ParseDNSPacket(packet)
	if err != nil {
		t.Fatalf("ParseDNSPacket error: %v", err)
	}

	if q.ID != 0x1234 {
		t.Errorf("Expected ID 0x1234, got 0x%04X", q.ID)
	}
	if q.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got %q", q.Domain)
	}
	if q.Type != "A" {
		t.Errorf("Expected Type 'A', got %q", q.Type)
	}
	if q.IsResponse {
		t.Errorf("Expected IsResponse=false for query")
	}
}

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer(3)

	q1 := &model.DNSQuery{Domain: "one.com", Timestamp: time.Now()}
	q2 := &model.DNSQuery{Domain: "two.com", Timestamp: time.Now()}
	q3 := &model.DNSQuery{Domain: "three.com", Timestamp: time.Now()}
	q4 := &model.DNSQuery{Domain: "four.com", Timestamp: time.Now()}

	rb.Add(q1)
	rb.Add(q2)
	rb.Add(q3)
	rb.Add(q4) // Evicts q1

	all := rb.GetAll()
	if len(all) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(all))
	}
	// Newest first
	if all[0].Domain != "four.com" {
		t.Errorf("Expected newest 'four.com', got %s", all[0].Domain)
	}
	if all[2].Domain != "two.com" {
		t.Errorf("Expected oldest retained 'two.com', got %s", all[2].Domain)
	}
}
