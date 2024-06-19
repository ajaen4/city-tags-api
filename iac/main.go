package main

import (
	"city-tags-api-iac/internal/config"
	"city-tags-api-iac/internal/containers"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.Load(ctx)

		services := containers.NewServices(cfg)
		services.Deploy()

		return nil
	})
}
