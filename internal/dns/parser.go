package dns

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/plxor/netpulse/internal/model"
)

var typeNames = map[uint16]string{
	1:   "A",
	2:   "NS",
	5:   "CNAME",
	6:   "SOA",
	12:  "PTR",
	15:  "MX",
	16:  "TXT",
	28:  "AAAA",
	33:  "SRV",
	65:  "HTTPS",
	255: "ANY",
}

var rcodeNames = map[uint8]string{
	0: "NOERROR",
	1: "FORMERR",
	2: "SERVFAIL",
	3: "NXDOMAIN",
	4: "NOTIMP",
	5: "REFUSED",
}

// ParseDNSPacket decodes a raw wire-format DNS packet into a model.DNSQuery.
func ParseDNSPacket(data []byte) (*model.DNSQuery, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("packet too short (%d bytes)", len(data))
	}

	txID := binary.BigEndian.Uint16(data[0:2])
	flags := binary.BigEndian.Uint16(data[2:4])
	qdCount := binary.BigEndian.Uint16(data[4:6])
	anCount := binary.BigEndian.Uint16(data[6:8])

	isResponse := (flags & 0x8000) != 0
	rcode := uint8(flags & 0x000F)

	rcodeStr := rcodeNames[rcode]
	if rcodeStr == "" {
		rcodeStr = fmt.Sprintf("RCODE_%d", rcode)
	}

	offset := 12
	var domain string
	var qtypeStr string

	// Parse Questions
	for i := 0; i < int(qdCount); i++ {
		name, nextOffset, err := readDomainName(data, offset)
		if err != nil {
			return nil, err
		}
		offset = nextOffset
		if offset+4 > len(data) {
			return nil, fmt.Errorf("truncated question")
		}
		qtype := binary.BigEndian.Uint16(data[offset : offset+2])
		// qclass := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4

		if domain == "" {
			domain = name
			qtypeStr = typeNames[qtype]
			if qtypeStr == "" {
				qtypeStr = fmt.Sprintf("TYPE_%d", qtype)
			}
		}
	}

	var answers []string
	// Parse Answers if present
	if isResponse {
		for i := 0; i < int(anCount); i++ {
			if offset >= len(data) {
				break
			}
			_, nextOffset, err := readDomainName(data, offset)
			if err != nil {
				break
			}
			offset = nextOffset
			if offset+10 > len(data) {
				break
			}

			atype := binary.BigEndian.Uint16(data[offset : offset+2])
			// aclass := binary.BigEndian.Uint16(data[offset+2 : offset+4])
			// ttl := binary.BigEndian.Uint32(data[offset+4 : offset+8])
			rdlen := int(binary.BigEndian.Uint16(data[offset+8 : offset+10]))
			offset += 10

			if offset+rdlen > len(data) {
				break
			}

			rdata := data[offset : offset+rdlen]
			offset += rdlen

			switch atype {
			case 1: // A
				if len(rdata) == 4 {
					answers = append(answers, net.IP(rdata).String())
				}
			case 28: // AAAA
				if len(rdata) == 16 {
					answers = append(answers, net.IP(rdata).String())
				}
			case 5: // CNAME
				cname, _, err := readDomainName(data, offset-rdlen)
				if err == nil && cname != "" {
					answers = append(answers, cname)
				}
			}
		}
	}

	query := &model.DNSQuery{
		ID:         txID,
		Domain:     domain,
		Type:       qtypeStr,
		RCode:      rcodeStr,
		Answers:    answers,
		IsResponse: isResponse,
	}

	return query, nil
}

func readDomainName(data []byte, offset int) (string, int, error) {
	var labels []string
	jumped := false
	startOffset := offset
	jumpOffset := 0
	visited := make(map[int]bool)

	for {
		if offset >= len(data) {
			return "", 0, fmt.Errorf("offset out of bounds")
		}

		length := int(data[offset])

		// Pointer check (11000000 = 0xC0)
		if (length & 0xC0) == 0xC0 {
			if offset+1 >= len(data) {
				return "", 0, fmt.Errorf("pointer out of bounds")
			}
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			if visited[pointer] {
				return "", 0, fmt.Errorf("pointer loop detected")
			}
			visited[pointer] = true

			if !jumped {
				jumpOffset = offset + 2
				jumped = true
			}
			offset = pointer
			continue
		}

		offset++
		if length == 0 {
			break
		}

		if offset+length > len(data) {
			return "", 0, fmt.Errorf("label length out of bounds")
		}

		labels = append(labels, string(data[offset:offset+length]))
		offset += length
	}

	finalOffset := offset
	if jumped {
		finalOffset = jumpOffset
	} else if finalOffset == startOffset {
		finalOffset = startOffset + 1
	}

	return strings.Join(labels, "."), finalOffset, nil
}
