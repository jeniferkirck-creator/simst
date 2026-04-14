package service

import "client/internal/models"

type Stresser interface {
	MakeRequestAndSend(request *models.Target) (uint16, int, error)
}
