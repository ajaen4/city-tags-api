package config

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Config struct {
	Region string
}

func Load(ctx *pulumi.Context) (*Config, error) {
	cfg := config.New(ctx, "")
	return &Config{
		Region: cfg.Require("region"),
	}, nil
}
