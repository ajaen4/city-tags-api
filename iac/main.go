package main

import (
	"city-tags-api-iac/internal/config"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg, err := config.Load(ctx)
		if err != nil {
			return err
		}
		fmt.Print(cfg)
		return nil
	})
}
