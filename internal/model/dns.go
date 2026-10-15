package model

import (
	"fmt"
	"strings"
	"time"
)

// DNSQuery represents a captured DNS lookup event.
type DNSQuery struct {
	ID         uint16
	Timestamp  time.Time
	ClientIP   string
	ClientPort int
	ServerIP   string
	ServerPort int
	Domain     string
	Type       string // "A", "AAAA", "CNAME", "PTR", "MX", "TXT", "HTTPS"
	RCode      string // "NOERROR", "NXDOMAIN", "SERVFAIL", "REFUSED"
	Answers    []string
	Latency    time.Duration
	IsResponse bool
	Process    string // Optional matched process name if known
}

// AnswerSummary formats answers into a compact readable string.
func (q *DNSQuery) AnswerSummary() string {
	if len(q.Answers) == 0 {
		if q.RCode != "" && q.RCode != "NOERROR" {
			return q.RCode
		}
		return "-"
	}
	if len(q.Answers) <= 2 {
		return strings.Join(q.Answers, ", ")
	}
	return fmt.Sprintf("%s, %s (+%d more)", q.Answers[0], q.Answers[1], len(q.Answers)-2)
}
