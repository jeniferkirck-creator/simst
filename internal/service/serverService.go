package service

import (
	"client/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ServerService struct {
	RegisterPath      string
	ConfigurationPath string
}

func NewServerService(registerPath, configurationPath string) *ServerService {
	return &ServerService{
		RegisterPath:      registerPath,
		ConfigurationPath: configurationPath,
	}
}

func (s *ServerService) Register(myIP string) error {
	if myIP == "" {
		var retry = 0
		for {
			time.Sleep(time.Duration(retry) * time.Second)
			ip, err := getMyIP()
			if err != nil {
				retry++
				continue
			}
			myIP = ip
			break
		}
	}
	r, err := http.Get(s.RegisterPath + fmt.Sprintf("?ip=%s", myIP))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode >= 400 {
		return errors.New(r.Status)
	}
	return nil
}

func (s *ServerService) Configuration() ([]*models.Target, error) {
	r, err := http.Get(s.ConfigurationPath)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}

	var targets []*models.Target
	if err = json.NewDecoder(r.Body).Decode(&targets); err != nil {
		return nil, err
	}
	return targets, nil
}

func getMyIP() (string, error) {
	r, err := http.Get("https://checkip.global.api.aws/")
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
