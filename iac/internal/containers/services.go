package containers

import (
	"city-tags-api-iac/internal/input"
)

type Services struct {
	cfg *input.Input
}

func NewServices(cfg *input.Input) *Services {
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
