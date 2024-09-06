package containers

import (
	"city-tags-api-iac/internal/config"
)

type Services struct {
	cfg *config.Config
}

func NewServices(cfg *config.Config) *Services {
	return &Services{
		cfg: cfg,
	}
}

func (servs *Services) Deploy() {
	for servName, servCfg := range servs.cfg.ServicesCfg {
		service := NewService(servs.cfg.Ctx, servName, servCfg)
		service.deploy()
	}
}
