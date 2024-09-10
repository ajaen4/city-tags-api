package main

import (
	"city-tags-api-iac/internal/containers"
	"city-tags-api-iac/internal/input"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := input.Load(ctx)

		containers.NewServices(cfg)
		containers.NewFunctions(cfg)

		return nil
	})
}
