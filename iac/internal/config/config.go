package config

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Config struct {
	Ctx         *pulumi.Context
	ServicesCfg map[string]*ServiceCfg
}

type ServiceCfg struct {
	BuildVersion  string `json:"build_version"`
	Cpu           int    `json:"cpu"`
	Memory        int    `json:"memory"`
	MinCount      int    `json:"min_count"`
	MaxCount      int    `json:"max_count"`
	LbPort        int    `json:"lb_port"`
	ContainerPort int    `json:"container_port"`
}

func Load(ctx *pulumi.Context) *Config {
	cfg := config.New(ctx, "")

	servicesCfg := map[string]*ServiceCfg{}
	cfg.RequireObject("services", &servicesCfg)

	return &Config{
		Ctx:         ctx,
		ServicesCfg: servicesCfg,
	}
}
