package containers

import (
	"city-tags-api-iac/internal/aws_lib"
	"city-tags-api-iac/internal/config"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/appautoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type service struct {
	ctx              *pulumi.Context
	name             string
	cfg              *config.ServiceCfg
	publicSubnetsOut pulumi.AnyOutput
	roles            map[string]*iam.Role
	sg               *ec2.SecurityGroup
	targetGroup      *lb.TargetGroup
}

func NewService(ctx *pulumi.Context, name string, servCfg *config.ServiceCfg, roles map[string]*iam.Role) *service {
	baselineNetRef, err := pulumi.NewStackReference(ctx, "ajaen4/sityex-baseline/main", nil)
	if err != nil {
		log.Fatal(err)
	}
	return &service{
		ctx:              ctx,
		name:             name,
		cfg:              servCfg,
		publicSubnetsOut: baselineNetRef.GetOutput(pulumi.String("public_subnet_ids")),
		roles:            roles,
	}
}

func (service *service) deploy() {
	service.createNetworking()
	service.createECSService()
}

func (service *service) createNetworking() {
	lbSGName := fmt.Sprintf("%s-lb-sg-%s", service.name, service.ctx.Stack())
	lbSG, err := ec2.NewSecurityGroup(
		service.ctx,
		lbSGName,
		&ec2.SecurityGroupArgs{
			Name:        pulumi.String(lbSGName),
			Description: pulumi.String("Controls access to the ALB"),
			VpcId:       pulumi.String("vpc-056a5820b4dc966b9"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(lbSGName),
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = vpc.NewSecurityGroupIngressRule(
		service.ctx,
		fmt.Sprintf("%s-lb-ingress-%s", service.name, service.ctx.Stack()),
		&vpc.SecurityGroupIngressRuleArgs{
			SecurityGroupId: lbSG.ID(),
			CidrIpv4:        pulumi.String("0.0.0.0/0"),
			FromPort:        pulumi.Int(service.cfg.LbPort),
			ToPort:          pulumi.Int(service.cfg.LbPort),
			IpProtocol:      pulumi.String("tcp"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = vpc.NewSecurityGroupEgressRule(
		service.ctx,
		fmt.Sprintf("%s-lb-egress-%s", service.name, service.ctx.Stack()),
		&vpc.SecurityGroupEgressRuleArgs{
			SecurityGroupId: lbSG.ID(),
			CidrIpv4:        pulumi.String("0.0.0.0/0"),
			FromPort:        pulumi.Int(0),
			ToPort:          pulumi.Int(0),
			IpProtocol:      pulumi.String("-1"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	srvSGName := fmt.Sprintf("%s-service-sg-%s", service.name, service.ctx.Stack())
	service.sg, err = ec2.NewSecurityGroup(
		service.ctx,
		srvSGName,
		&ec2.SecurityGroupArgs{
			Name:        pulumi.String(srvSGName),
			Description: pulumi.String("Controls access to the ECS Service"),
			VpcId:       pulumi.String("vpc-056a5820b4dc966b9"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(srvSGName),
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = vpc.NewSecurityGroupIngressRule(
		service.ctx,
		fmt.Sprintf("%s-service-ingress-%s", service.name, service.ctx.Stack()),
		&vpc.SecurityGroupIngressRuleArgs{
			SecurityGroupId:           service.sg.ID(),
			FromPort:                  pulumi.Int(0),
			ToPort:                    pulumi.Int(0),
			IpProtocol:                pulumi.String("-1"),
			ReferencedSecurityGroupId: lbSG.ID(),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = vpc.NewSecurityGroupEgressRule(
		service.ctx,
		fmt.Sprintf("%s-service-egress-%s", service.name, service.ctx.Stack()),
		&vpc.SecurityGroupEgressRuleArgs{
			SecurityGroupId: service.sg.ID(),
			CidrIpv4:        pulumi.String("0.0.0.0/0"),
			FromPort:        pulumi.Int(0),
			ToPort:          pulumi.Int(0),
			IpProtocol:      pulumi.String("-1"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	lbName := fmt.Sprintf("%s-lb-%s", service.name, service.ctx.Stack())
	serviceLB, err := lb.NewLoadBalancer(
		service.ctx,
		lbName,
		&lb.LoadBalancerArgs{
			Name:             pulumi.String(lbName),
			LoadBalancerType: pulumi.String("application"),
			SecurityGroups:   pulumi.StringArray{lbSG.ID()},
			SubnetMappings: lb.LoadBalancerSubnetMappingArray{
				&lb.LoadBalancerSubnetMappingArgs{
					SubnetId: service.publicSubnetsOut.ApplyT(func(id any) string {
						return id.([]any)[0].(string)
					}).(pulumi.StringOutput),
				},
				&lb.LoadBalancerSubnetMappingArgs{
					SubnetId: service.publicSubnetsOut.ApplyT(func(id any) string {
						return id.([]any)[1].(string)
					}).(pulumi.StringOutput),
				},
			},
			Internal: pulumi.Bool(false),
		})
	if err != nil {
		log.Fatal(err)
	}

	TargGrName := fmt.Sprintf("%s-target-group-%s", service.name, service.ctx.Stack())
	service.targetGroup, err = lb.NewTargetGroup(
		service.ctx,
		TargGrName,
		&lb.TargetGroupArgs{
			Name:       pulumi.String(TargGrName),
			Port:       pulumi.Int(service.cfg.LbPort),
			Protocol:   pulumi.String("HTTP"),
			VpcId:      pulumi.String("vpc-056a5820b4dc966b9"),
			TargetType: pulumi.String("ip"),
			HealthCheck: lb.TargetGroupHealthCheckArgs{
				Path:               pulumi.String("/ping"),
				Port:               pulumi.String("traffic-port"),
				HealthyThreshold:   pulumi.IntPtr(5),
				UnhealthyThreshold: pulumi.IntPtr(2),
				Timeout:            pulumi.IntPtr(2),
				Interval:           pulumi.IntPtr(5),
				Matcher:            pulumi.String("200"),
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = lb.NewListener(
		service.ctx,
		fmt.Sprintf("%s-listener-%s", service.name, service.ctx.Stack()),
		&lb.ListenerArgs{
			LoadBalancerArn: serviceLB.Arn,
			Port:            pulumi.Int(service.cfg.LbPort),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: service.targetGroup.Arn,
				},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

}

func (service *service) createECSService() {
	ecrRepo := NewRepository(
		service.ctx,
		fmt.Sprintf("%s-repository-%s", service.name, service.ctx.Stack()),
	)
	image := NewImage(
		service.ctx,
		fmt.Sprintf("%s-image-%s", service.name, service.ctx.Stack()),
		ecrRepo.EcrRepository,
	)
	imageURI := image.PushImage(service.cfg.BuildVersion)

	logGroup, err := cloudwatch.NewLogGroup(
		service.ctx,
		fmt.Sprintf("%s-log-group-%s", service.name, service.ctx.Stack()),
		&cloudwatch.LogGroupArgs{
			Name:            pulumi.String(fmt.Sprintf("ecs/%s", service.name)),
			RetentionInDays: pulumi.Int(30),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	clusterName := fmt.Sprintf("%s-cluster-%s", service.name, service.ctx.Stack())
	cluster, err := ecs.NewCluster(
		service.ctx,
		clusterName,
		&ecs.ClusterArgs{
			Name: pulumi.String(clusterName),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	envVars, err := json.Marshal(service.getEnvVars())
	if err != nil {
		log.Fatal(err)
	}

	containerDef := pulumi.All(imageURI, logGroup.Name).ApplyT(
		func(args []any) pulumi.StringOutput {
			return pulumi.Sprintf(`[
				{
					"name": "%s",
					"image": "%s",
					"essential": true,
					"cpu": %d,
					"memory": %d,
					"entryPoint": ["./main"],
					"portMappings": [
						{
							"containerPort": %d,
							"protocol": "tcp"
						}
					],
					"logConfiguration": {
						"logDriver": "awslogs",
						"options": {
							"awslogs-group": "%s",
							"awslogs-region": "eu-west-1",
							"awslogs-stream-prefix": "%s-log-stream"
						}
					},
					"environment": %s
				}
			]`,
				service.name,
				args[0],
				service.cfg.Cpu,
				service.cfg.Memory,
				service.cfg.ContainerPort,
				args[1],
				service.name,
				envVars,
			)
		},
	).(pulumi.StringOutput)

	taskDefName := fmt.Sprintf("%s-task-def-%s", service.name, service.ctx.Stack())
	taskDef, err := ecs.NewTaskDefinition(
		service.ctx,
		taskDefName,
		&ecs.TaskDefinitionArgs{
			Family:                  pulumi.String(taskDefName),
			NetworkMode:             pulumi.String("awsvpc"),
			ContainerDefinitions:    containerDef,
			RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
			Cpu:                     pulumi.StringPtr(fmt.Sprint(service.cfg.Cpu)),
			Memory:                  pulumi.StringPtr(fmt.Sprint(service.cfg.Memory)),
			ExecutionRoleArn:        service.roles["exec_role"].Arn,
			TaskRoleArn:             service.roles["task_role"].Arn,
		})
	if err != nil {
		log.Fatal(taskDef)
	}

	serviceName := fmt.Sprintf("%s-service-%s", service.name, service.ctx.Stack())
	_, err = ecs.NewService(
		service.ctx,
		serviceName,
		&ecs.ServiceArgs{
			Name:           pulumi.String(serviceName),
			Cluster:        cluster.ID(),
			TaskDefinition: taskDef.Arn,
			DesiredCount:   pulumi.Int(service.cfg.MinCount),
			LaunchType:     pulumi.String("FARGATE"),
			LoadBalancers: ecs.ServiceLoadBalancerArray{
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: service.targetGroup.Arn,
					ContainerName:  pulumi.String(service.name),
					ContainerPort:  pulumi.Int(service.cfg.ContainerPort),
				},
			},
			NetworkConfiguration: ecs.ServiceNetworkConfigurationArgs{
				Subnets:        service.publicSubnetsOut.AsStringArrayOutput(),
				SecurityGroups: pulumi.StringArray{service.sg.ID()},
				AssignPublicIp: pulumi.Bool(true),
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	serviceId := pulumi.Sprintf(
		"service/%s/%s", clusterName, serviceName,
	)
	scalingTarget, err := appautoscaling.NewTarget(
		service.ctx,
		fmt.Sprintf("%s-scaling-target-%s", service.name, service.ctx.Stack()),
		&appautoscaling.TargetArgs{
			ResourceId:        serviceId,
			MinCapacity:       pulumi.Int(service.cfg.MinCount),
			MaxCapacity:       pulumi.Int(service.cfg.MaxCount),
			ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
			ServiceNamespace:  pulumi.String("ecs"),
		})
	if err != nil {
		log.Fatal(err)
	}

	_, err = appautoscaling.NewPolicy(
		service.ctx,
		fmt.Sprintf("%s-scaling-policy-%s", service.name, service.ctx.Stack()),
		&appautoscaling.PolicyArgs{
			PolicyType:        pulumi.String("TargetTrackingScaling"),
			ResourceId:        serviceId,
			ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
			ServiceNamespace:  pulumi.String("ecs"),
			TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
				TargetValue:      pulumi.Float64(70.0),
				ScaleInCooldown:  pulumi.Int(60),
				ScaleOutCooldown: pulumi.Int(60),
				PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
					PredefinedMetricType: pulumi.String("ECSServiceAverageCPUUtilization"),
				},
			},
		},
		pulumi.DependsOn([]pulumi.Resource{scalingTarget}),
	)
	if err != nil {
		log.Fatal(err)
	}

}

func (service *service) getEnvVars() []map[string]string {
	envVars := []map[string]string{}
	ssm := aws_lib.NewSSM()
	for _, envVarCfg := range service.cfg.EnvVars {
		if envVarCfg.Type == "SSM" {
			ssmEnvVars := ssm.GetParam(envVarCfg.Path, true)
			for name, value := range ssmEnvVars {
				envVars = append(envVars, map[string]string{"name": name, "value": value})
			}
		} else {
			envVars = append(envVars, ssm.GetParam(envVarCfg.Path, true))
		}
	}
	return envVars
}
