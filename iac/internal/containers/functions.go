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
	"github.com/pulumiverse/pulumi-time/sdk/go/time"
)

func NewFunctions(cfg *input.Input) {
	for funcName, funcCfg := range cfg.FunctionsCfg {
		NewFunction(cfg.Ctx, funcName, funcCfg)
	}
}

type function struct {
	ctx  *pulumi.Context
	name string
	cfg  *input.FunctionCfg
}

func NewFunction(ctx *pulumi.Context, name string, funcCfg *input.FunctionCfg) {
	function := &function{
		ctx:  ctx,
		name: name,
		cfg:  funcCfg,
	}
	function.deploy()
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
			Project: pulumi.String(input.GetProject()),
			Role:    pulumi.String("roles/run.invoker"),
			Member:  pulumi.Sprintf("serviceAccount:%s", sa.Email),
		})
	if err != nil {
		log.Fatal(err)
	}

	return sa
}

func (function *function) createFunction(sa *serviceaccount.Account) {
	repo := NewRepository(function.ctx, function.name, input.GetRegion())
	image := NewImage(
		function.ctx,
		function.cfg.ImgCfg,
		function.name,
		repo,
	)
	imageUrl := image.PushImage(function.cfg.BuildVersion)

	sleep, err := time.NewSleep(
		function.ctx,
		fmt.Sprintf("%s-sleep", function.name),
		&time.SleepArgs{
			CreateDuration: pulumi.String("5s"),
		},
		pulumi.DependsOn([]pulumi.Resource{image.Resource}),
	)
	if err != nil {
		log.Fatalf("Failed to create sleep resource: %v", err)
	}

	job, err := cloudrunv2.NewJob(
		function.ctx,
		function.name,
		&cloudrunv2.JobArgs{
			Name:     pulumi.String(function.name),
			Location: pulumi.String(input.GetRegion()),
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
		pulumi.DependsOn([]pulumi.Resource{sleep}),
	)
	if err != nil {
		function.ctx.Log.Error(fmt.Sprintf("Failed to create job: %v", err), nil)
		return
	}

	_, err = cloudscheduler.NewJob(
		function.ctx,
		fmt.Sprintf("%s-scheduler", function.name),
		&cloudscheduler.JobArgs{
			Name:     pulumi.String(fmt.Sprintf("%s-scheduler", function.name)),
			Project:  pulumi.String(input.GetProject()),
			Region:   pulumi.String(input.GetRegion()),
			Schedule: pulumi.String(function.cfg.ScheduleExp),
			TimeZone: pulumi.String("Etc/UTC"),
			HttpTarget: &cloudscheduler.JobHttpTargetArgs{
				HttpMethod: pulumi.String("POST"),
				Uri: pulumi.Sprintf("https://%s-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/%s/jobs/%s:run",
					input.GetRegion(),
					input.GetProject(),
					function.name,
				),
				OauthToken: &cloudscheduler.JobHttpTargetOauthTokenArgs{
					ServiceAccountEmail: sa.Email,
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{job}),
	)
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
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
