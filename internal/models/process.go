package models

type Process struct {
	Timestamp     int64  `json:"timestamp"`
	RequestsCount uint64 `json:"requests_count"`
	PayloadLength uint64 `json:"payload_length"`
	Codes         []int  `json:"codes"`
}
