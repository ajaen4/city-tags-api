package containers

import (
	"city-tags-api-iac/internal/input"
	"fmt"
	"log"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type function struct {
	ctx  *pulumi.Context
	name string
	cfg  *input.FunctionCfg
}

func NewFunction(ctx *pulumi.Context, name string, funcCfg *input.FunctionCfg) *function {
	return &function{
		ctx:  ctx,
		name: name,
		cfg:  funcCfg,
	}
}

func (function *function) deploy() {
	function.createFunction()
}

func (function *function) createFunction() {
	repo := NewRepository(function.ctx, function.name, function.cfg.Region)
	image := NewImage(
		function.ctx,
		function.cfg.ImgCfg,
		function.cfg.Project,
		function.cfg.Region,
		fmt.Sprintf("%s-%s", function.name, function.ctx.Stack()),
		repo,
	)
	_ = image.PushImage(function.cfg.BuildVersion)

	// function

	_, err := organizations.LookupIAMPolicy(
		function.ctx,
		&organizations.LookupIAMPolicyArgs{
			Bindings: []organizations.GetIAMPolicyBinding{
				{
					Role: "roles/run.invoker",
					Members: []string{
						"allUsers",
					},
				},
			},
		}, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (function *function) parseEnvs() cloudrun.ServiceTemplateSpecContainerEnvArray {
	var envs cloudrun.ServiceTemplateSpecContainerEnvArray
	for _, env := range function.cfg.EnvVars {
		env_var := &cloudrun.ServiceTemplateSpecContainerEnvArgs{
			Name:  pulumi.String(env.Name),
			Value: pulumi.String(env.Value),
		}
		envs = append(envs, env_var)
	}
	return envs
}
