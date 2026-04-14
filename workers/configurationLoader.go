package workers

import (
	"client/internal/models"
	"client/internal/service"
	"context"
	"encoding/json"
	"os"
	"time"
)

type ConfigurationLoader struct {
	loader        chan []*models.Target
	filePath      string
	serverService *service.ServerService
}

func NewConfigurationLoader(ch chan []*models.Target, filePath string, srv *service.ServerService) *ConfigurationLoader {
	return &ConfigurationLoader{
		loader:        ch,
		filePath:      filePath,
		serverService: srv,
	}
}

func (c *ConfigurationLoader) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if c.serverService != nil {
				list, err := loadFromServer(c.serverService)
				if err != nil {
					continue
				}
				c.loader <- list
			} else if c.filePath != "" {
				list, err := loadFromFile(c.filePath)
				if err != nil {
					continue
				}
				c.loader <- list
			}
		}
	}
}

func loadFromFile(path string) ([]*models.Target, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var targets []*models.Target
	if err = json.NewDecoder(f).Decode(&targets); err != nil {
		return nil, err
	}
	return targets, nil
}

func loadFromServer(srv *service.ServerService) ([]*models.Target, error) {
	cfg, err := srv.Configuration()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
