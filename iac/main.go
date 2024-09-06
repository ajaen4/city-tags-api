package main

import (
	"city-tags-api-iac/internal/containers"
	"city-tags-api-iac/internal/input"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := input.Load(ctx)

		services := containers.NewServices(cfg)
		services.Deploy()

		funcs := containers.NewFunctions(cfg)
		funcs.Deploy()

		return nil
	})
}
