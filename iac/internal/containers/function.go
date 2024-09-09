package containers

import (
	"city-tags-api-iac/internal/input"
	"fmt"
	"log"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/cloudscheduler"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
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
	sa := function.createServiceAccount()
	function.createFunction(sa)
}

func (function *function) createServiceAccount() *serviceaccount.Account {
	accountId := fmt.Sprintf("%s-sa", function.name)
	sa, err := serviceaccount.NewAccount(
		function.ctx,
		accountId,
		&serviceaccount.AccountArgs{
			AccountId:   pulumi.String(accountId),
			DisplayName: pulumi.String(fmt.Sprintf("Service Account for %s", function.name)),
		})
	if err != nil {
		log.Fatal(err)
	}

	_, err = projects.NewIAMMember(
		function.ctx,
		fmt.Sprintf("%s-invoker", function.name),
		&projects.IAMMemberArgs{
			Project: pulumi.String(function.cfg.Project),
			Role:    pulumi.String("roles/run.invoker"),
			Member:  pulumi.Sprintf("serviceAccount:%s", sa.Email),
		})
	if err != nil {
		log.Fatal(err)
	}

	return sa
}

func (function *function) createFunction(sa *serviceaccount.Account) {
	repo := NewRepository(function.ctx, function.name, function.cfg.Region)
	image := NewImage(
		function.ctx,
		function.cfg.ImgCfg,
		function.cfg.Project,
		function.cfg.Region,
		fmt.Sprintf("%s-%s", function.name, function.ctx.Stack()),
		repo,
	)
	imageUrl := image.PushImage(function.cfg.BuildVersion)

	_, err := cloudrunv2.NewJob(
		function.ctx,
		function.name,
		&cloudrunv2.JobArgs{
			Name:     pulumi.String(function.name),
			Location: pulumi.String(function.cfg.Region),
			Template: &cloudrunv2.JobTemplateArgs{
				Template: &cloudrunv2.JobTemplateTemplateArgs{
					Containers: cloudrunv2.JobTemplateTemplateContainerArray{
						&cloudrunv2.JobTemplateTemplateContainerArgs{
							Image:    imageUrl,
							Envs:     function.parseEnvs(),
							Commands: pulumi.ToStringArray(function.cfg.Entrypoint),
						},
					},
					ServiceAccount: sa.Email,
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{image.Resource}),
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = cloudscheduler.NewJob(
		function.ctx,
		fmt.Sprintf("%s-scheduler", function.name),
		&cloudscheduler.JobArgs{
			Name:     pulumi.String(fmt.Sprintf("%s-scheduler", function.name)),
			Project:  pulumi.String(function.cfg.Project),
			Region:   pulumi.String(function.cfg.Region),
			Schedule: pulumi.String("0 0 * * 0"),
			TimeZone: pulumi.String("Etc/UTC"),
			HttpTarget: &cloudscheduler.JobHttpTargetArgs{
				HttpMethod: pulumi.String("POST"),
				Uri: pulumi.Sprintf("https://%s-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/%s/jobs/%s:run",
					function.cfg.Region,
					function.cfg.Project,
					function.name,
				),
				OauthToken: &cloudscheduler.JobHttpTargetOauthTokenArgs{
					ServiceAccountEmail: sa.Email,
				},
			},
		})
	if err != nil {
		log.Fatal(err)
	}
}

func (function *function) parseEnvs() cloudrunv2.JobTemplateTemplateContainerEnvArray {
	var envs cloudrunv2.JobTemplateTemplateContainerEnvArray
	for _, env := range function.cfg.EnvVars {
		env_var := &cloudrunv2.JobTemplateTemplateContainerEnvArgs{
			Name:  pulumi.String(env.Name),
			Value: pulumi.String(env.Value),
		}
		envs = append(envs, env_var)
	}
	return envs
}
