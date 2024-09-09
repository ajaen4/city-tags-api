package containers

import (
	"fmt"
	"log"

	"city-tags-api-iac/internal/input"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/route53"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/secretmanager"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type service struct {
	ctx  *pulumi.Context
	name string
	cfg  *input.ServiceCfg
}

func NewService(ctx *pulumi.Context, name string, servCfg *input.ServiceCfg) *service {
	return &service{
		ctx:  ctx,
		name: name,
		cfg:  servCfg,
	}
}

func (service *service) deploy() {
	sa := service.createServiceAccount()
	service.createService(sa)
}

func (service *service) createServiceAccount() *serviceaccount.Account {
	accountId := fmt.Sprintf("%s-sa", service.name)
	sa, err := serviceaccount.NewAccount(
		service.ctx,
		accountId,
		&serviceaccount.AccountArgs{
			AccountId:   pulumi.String(accountId),
			DisplayName: pulumi.String("Service Account for accessing secrets"),
		})
	if err != nil {
		log.Fatal(err)
	}

	secretNames := []string{
		"city-tags-api-db",
		"city-tags-api-secret",
	}

	member := sa.Email.ApplyT(func(email string) string {
		return fmt.Sprintf("serviceAccount:%s", email)
	}).(pulumi.StringInput)

	for _, secretName := range secretNames {
		_, err := secretmanager.NewSecretIamMember(
			service.ctx,
			fmt.Sprintf("%s-access", secretName),
			&secretmanager.SecretIamMemberArgs{
				SecretId: pulumi.String(secretName),
				Role:     pulumi.String("roles/secretmanager.secretAccessor"),
				Member:   member,
			})
		if err != nil {
			log.Fatal(err)
		}
	}
	return sa
}

func (service *service) createService(sa *serviceaccount.Account) {
	repo := NewRepository(service.ctx, service.name, service.cfg.Region)
	image := NewImage(
		service.ctx,
		service.cfg.ImgCfg,
		service.cfg.Project,
		service.cfg.Region,
		fmt.Sprintf("%s-%s", service.name, service.ctx.Stack()),
		repo,
	)
	imageUrl := image.PushImage(service.cfg.BuildVersion)

	crService, err := cloudrun.NewService(
		service.ctx,
		service.name,
		&cloudrun.ServiceArgs{
			Name:     pulumi.String(service.name),
			Location: pulumi.String(service.cfg.Region),
			Template: &cloudrun.ServiceTemplateArgs{
				Spec: &cloudrun.ServiceTemplateSpecArgs{
					ServiceAccountName: sa.Email,
					Containers: cloudrun.ServiceTemplateSpecContainerArray{
						&cloudrun.ServiceTemplateSpecContainerArgs{
							Image:    imageUrl,
							Envs:     service.parseEnvs(),
							Commands: pulumi.ToStringArray(service.cfg.Entrypoint),
							Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
								&cloudrun.ServiceTemplateSpecContainerPortArgs{
									ContainerPort: pulumi.Int(service.cfg.ContainerPort),
								},
							},
							Resources: &cloudrun.ServiceTemplateSpecContainerResourcesArgs{
								Limits: pulumi.StringMap{
									"cpu": pulumi.Sprintf("%d", service.cfg.Cpu*2),
								},
								Requests: pulumi.StringMap{
									"cpu":    pulumi.Sprintf("%d", service.cfg.Cpu),
									"memory": pulumi.String(service.cfg.Memory),
								},
							},
							StartupProbe: &cloudrun.ServiceTemplateSpecContainerStartupProbeArgs{
								HttpGet: &cloudrun.ServiceTemplateSpecContainerStartupProbeHttpGetArgs{
									Port: pulumi.Int(service.cfg.ContainerPort),
									Path: pulumi.String("/ping"),
								},
							},
						},
					},
				},
				Metadata: &cloudrun.ServiceTemplateMetadataArgs{
					Annotations: pulumi.StringMap{
						"autoscaling.knative.dev/maxScale": pulumi.Sprintf("%d", service.cfg.MaxCount),
						"autoscaling.knative.dev/minScale": pulumi.Sprintf("%d", service.cfg.MinCount),
					},
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{image.Resource}),
	)
	if err != nil {
		log.Fatal(err)
	}

	noauth, err := organizations.LookupIAMPolicy(
		service.ctx,
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

	_, err = cloudrun.NewIamPolicy(
		service.ctx,
		"no-auth",
		&cloudrun.IamPolicyArgs{
			Location:   pulumi.String(service.cfg.Region),
			Service:    crService.Name,
			PolicyData: pulumi.String(noauth.PolicyData),
		})
	if err != nil {
		log.Fatal(err)
	}

	domainMapping := service.createDomainMapping(crService, "city-tags-api.com")
	service.createDNSRecords(domainMapping)
}

func (service *service) createDomainMapping(crService *cloudrun.Service, domain string) *cloudrun.DomainMapping {
	mapping, err := cloudrun.NewDomainMapping(
		service.ctx,
		fmt.Sprintf("%s-mapping", domain),
		&cloudrun.DomainMappingArgs{
			Name:     pulumi.String(domain),
			Location: crService.Location,
			Metadata: &cloudrun.DomainMappingMetadataArgs{
				Namespace: pulumi.String(service.cfg.Project),
			},
			Spec: &cloudrun.DomainMappingSpecArgs{
				RouteName: crService.Name,
			},
		})
	if err != nil {
		log.Fatal(err)
	}

	return mapping
}

func (service *service) createDNSRecords(domainMapping *cloudrun.DomainMapping) {
	ipAddress := domainMapping.Statuses.ApplyT(func(statuses []cloudrun.DomainMappingStatus) string {
		if len(statuses) > 0 && len(statuses[0].ResourceRecords) > 0 {
			return *statuses[0].ResourceRecords[0].Rrdata
		}
		return ""
	}).(pulumi.StringOutput)

	_, err := route53.NewRecord(
		service.ctx,
		"city-tags-api-root-a",
		&route53.RecordArgs{
			ZoneId:  pulumi.String(service.cfg.HostedZoneId),
			Name:    pulumi.String(""),
			Type:    pulumi.String("A"),
			Records: pulumi.StringArray{ipAddress},
			Ttl:     pulumi.Int(300),
		})
	if err != nil {
		log.Fatal(err)
	}

	service.ctx.Export("cloudRunIPAddress", ipAddress)
}

func (service *service) parseEnvs() cloudrun.ServiceTemplateSpecContainerEnvArray {
	var envs cloudrun.ServiceTemplateSpecContainerEnvArray
	for _, env := range service.cfg.EnvVars {
		env_var := &cloudrun.ServiceTemplateSpecContainerEnvArgs{
			Name:  pulumi.String(env.Name),
			Value: pulumi.String(env.Value),
		}
		envs = append(envs, env_var)
	}
	return envs
}
