//go:build windows

package service

import (
	"client/internal/models"
)

func init() {}

func (s *LevelFourStressService) MakeRequestAndSend(request *models.Target) (uint16, int, error) {
	return 0, 0, nil
}

func TcpSocketClient() int {
	return 0
}

func UdpSocketClient() int {
	return 0
}
