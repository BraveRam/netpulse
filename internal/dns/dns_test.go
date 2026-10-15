package dns

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/BraveRam/netpulse/internal/model"
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

func testDNSFrame(response bool, clientPort uint16) []byte {
	payload := []byte{0x12, 0x34, 0x01, 0, 0, 1, 0, 0, 0, 0, 0, 0, 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, 0, 1, 0, 1}
	if response {
		payload[2] = 0x81
	}
	packet := make([]byte, 28+len(payload))
	packet[0], packet[9] = 0x45, 17
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)))
	src, dst, sp, dp := net.ParseIP("192.168.1.2").To4(), net.ParseIP("8.8.8.8").To4(), clientPort, uint16(53)
	if response {
		src, dst, sp, dp = dst, src, dp, sp
	}
	copy(packet[12:16], src)
	copy(packet[16:20], dst)
	binary.BigEndian.PutUint16(packet[20:22], sp)
	binary.BigEndian.PutUint16(packet[22:24], dp)
	binary.BigEndian.PutUint16(packet[24:26], uint16(8+len(payload)))
	copy(packet[28:], payload)
	return packet
}

func TestDNSResponseEndpointsAndConcurrentTransactions(t *testing.T) {
	s := NewSniffer(10)
	defer s.Stop()
	s.processPacket(testDNSFrame(false, 40001))
	s.processPacket(testDNSFrame(false, 40002))
	if len(s.pendingTimes) != 2 {
		t.Fatalf("concurrent requests collided: %d pending", len(s.pendingTimes))
	}
	// Distinct known ages verify each reply consumes its own timestamp.
	for k := range s.pendingTimes {
		age := time.Second
		if k.clientPort == 40002 {
			age = 10 * time.Second
		}
		s.pendingTimes[k] = time.Now().Add(-age)
	}
	for _, port := range []uint16{40001, 40002} {
		s.processPacket(testDNSFrame(true, port))
		q := s.GetQueries()[0]
		if q.ClientIP != "192.168.1.2" || q.ClientPort != int(port) || q.ServerIP != "8.8.8.8" || q.ServerPort != 53 {
			t.Fatalf("wrong response endpoints: %+v", q)
		}
		min, max := time.Second, 5*time.Second
		if port == 40002 {
			min, max = 10*time.Second, 15*time.Second
		}
		if q.Latency < min || q.Latency >= max {
			t.Fatalf("wrong transaction latency: %v", q.Latency)
		}
	}
	if len(s.pendingTimes) != 0 {
		t.Fatal("completed requests remain pending")
	}
}
