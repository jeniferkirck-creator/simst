package service

import (
	"net/http"
	"sync"
)

type LevelSevenStressService struct {
	clientHTTP1, clientHTTP2, clientHTTP3 *http.Client
	request                               *http.Request
	mu                                    sync.RWMutex
}

func NewLevelSevenStressService(http1, http2, http3 *http.Client) Stresser {
	return &LevelSevenStressService{
		clientHTTP1: http1,
		clientHTTP2: http2,
		clientHTTP3: http3,
	}
}

type LevelFourStressService struct {
	clientTCP, clientUDP int
	payload              []byte
	mu                   sync.RWMutex
	TargetIP             string
}

func NewLevelFourStressService(tcp, udp int) Stresser {
	return &LevelFourStressService{
		clientTCP: tcp,
		clientUDP: udp,
	}
}
