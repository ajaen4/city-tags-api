package config

import (
	"log"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Config struct {
	Ctx         *pulumi.Context
	ServicesCfg map[string]*ServiceCfg
}

type EnvVar struct {
	Type  string `json:"type"`
	Path  string `json:"path"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ServiceCfg struct {
	BuildVersion  string   `json:"build_version"`
	Cpu           int      `json:"cpu"`
	Memory        int      `json:"memory"`
	MinCount      int      `json:"min_count"`
	MaxCount      int      `json:"max_count"`
	LbPort        int      `json:"lb_port"`
	ContainerPort int      `json:"container_port"`
	EnvVars       []EnvVar `json:"env_vars"`
}

func Load(ctx *pulumi.Context) *Config {
	cfg := config.New(ctx, "")

	servicesCfg := map[string]*ServiceCfg{}
	cfg.RequireObject("services", &servicesCfg)
	Validate(servicesCfg)

	return &Config{
		Ctx:         ctx,
		ServicesCfg: servicesCfg,
	}
}

func Validate(servicesCfg map[string]*ServiceCfg) {
	for _, sCfg := range servicesCfg {
		for _, envVar := range sCfg.EnvVars {
			if envVar.Type == "SSM" && envVar.Path == "" {
				log.Fatalf("Config validation error: incorrect env var, %s", envVar)
			}
			if envVar.Type == "" && (envVar.Name == "" || envVar.Value == "") {
				log.Fatalf("Config validation error: incorrect env var, %s", envVar)
			}
		}
	}
}
