package models

import "net/http"

type TargetType uint16

func (t TargetType) Uint16() uint16 {
	return uint16(t)
}

const (
	TCP TargetType = 1 << iota
	UDP
	HTTP1
	HTTP2
	HTTP3
	GET
	POST
	PUT
	DELETE
)

type Target struct {
	TargetHost                string            `json:"target_host"`
	TargetPort                uint16            `json:"target_port"`
	SourcePort                uint16            `json:"source_port,omitempty"`
	SourceIP                  string            `json:"source_ip,omitempty"`
	TargetIP                  string            `json:"target_ip,omitempty"`
	TargetQuery               string            `json:"target_query,omitempty"`
	Headers                   http.Header       `json:"headers,omitempty"`
	Cookies                   map[string]string `json:"cookies,omitempty"`
	Type                      TargetType        `json:"type"`
	Payload                   string            `json:"payload,omitempty"`
	WithRandomizer            bool              `json:"with_randomizer"`
	WithRandomAgent           bool              `json:"with_random_agent"`
	WithRandomSourceIPAndPort bool              `json:"with_random_source_ip_and_port"`
	PacketSequenceNumber      uint32            `json:"packet_sequence_number,omitempty"`
}
