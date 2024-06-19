package containers

import (
	"city-tags-api-iac/internal/config"
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
	servCfg          *config.ServiceCfg
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
		servCfg:          servCfg,
		publicSubnetsOut: baselineNetRef.GetOutput(pulumi.String("public_subnet_ids")),
		roles:            roles,
	}
}

func (service *service) deploy() {
	service.createNetworking()
	service.createECSService()
}

func (service *service) createNetworking() {
	lbSGName := fmt.Sprintf("%s-lb-sg", service.name)
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
		fmt.Sprintf("%s-lb-ingress", service.name),
		&vpc.SecurityGroupIngressRuleArgs{
			SecurityGroupId: lbSG.ID(),
			CidrIpv4:        pulumi.String("0.0.0.0/0"),
			FromPort:        pulumi.Int(service.servCfg.LbPort),
			ToPort:          pulumi.Int(service.servCfg.LbPort),
			IpProtocol:      pulumi.String("tcp"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = vpc.NewSecurityGroupEgressRule(
		service.ctx,
		fmt.Sprintf("%s-lb-egress", service.name),
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

	srvSGName := fmt.Sprintf("%s-service-sg", service.name)
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
		fmt.Sprintf("%s-service-ingress", service.name),
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
		fmt.Sprintf("%s-service-egress", service.name),
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

	lbName := fmt.Sprintf("%s-lb", service.name)
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

	TargGrName := fmt.Sprintf("%s-target-group", service.name)
	service.targetGroup, err = lb.NewTargetGroup(
		service.ctx,
		TargGrName,
		&lb.TargetGroupArgs{
			Name:       pulumi.String(TargGrName),
			Port:       pulumi.Int(service.servCfg.LbPort),
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
		fmt.Sprintf("%s-listener", service.name),
		&lb.ListenerArgs{
			LoadBalancerArn: serviceLB.Arn,
			Port:            pulumi.Int(service.servCfg.LbPort),
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
		fmt.Sprintf("%s-repository", service.name),
	)
	image := NewImage(
		service.ctx,
		fmt.Sprintf("%s-image", service.name),
		ecrRepo.EcrRepository,
	)
	imageURI := image.PushImage(service.servCfg.BuildVersion)

	logGroup, err := cloudwatch.NewLogGroup(
		service.ctx,
		fmt.Sprintf("%s-log-group", service.name),
		&cloudwatch.LogGroupArgs{
			Name:            pulumi.String(fmt.Sprintf("ecs/%s", service.name)),
			RetentionInDays: pulumi.Int(30),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	clusterName := fmt.Sprintf("%s-cluster", service.name)
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
					}
				}
			]`,
				service.name,
				args[0],
				service.servCfg.Cpu,
				service.servCfg.Memory,
				service.servCfg.ContainerPort,
				args[1],
				service.name,
			)
		},
	).(pulumi.StringOutput)

	taskDefName := fmt.Sprintf("%s-task-def", service.name)
	taskDef, err := ecs.NewTaskDefinition(
		service.ctx,
		taskDefName,
		&ecs.TaskDefinitionArgs{
			Family:                  pulumi.String(taskDefName),
			NetworkMode:             pulumi.String("awsvpc"),
			ContainerDefinitions:    containerDef,
			RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
			Cpu:                     pulumi.StringPtr(fmt.Sprint(service.servCfg.Cpu)),
			Memory:                  pulumi.StringPtr(fmt.Sprint(service.servCfg.Memory)),
			ExecutionRoleArn:        service.roles["exec_role"].Arn,
			TaskRoleArn:             service.roles["task_role"].Arn,
		})
	if err != nil {
		log.Fatal(taskDef)
	}

	serviceName := fmt.Sprintf("%s-service", service.name)
	_, err = ecs.NewService(
		service.ctx,
		serviceName,
		&ecs.ServiceArgs{
			Name:           pulumi.String(serviceName),
			Cluster:        cluster.ID(),
			TaskDefinition: taskDef.Arn,
			DesiredCount:   pulumi.Int(service.servCfg.MinCount),
			LaunchType:     pulumi.String("FARGATE"),
			LoadBalancers: ecs.ServiceLoadBalancerArray{
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: service.targetGroup.Arn,
					ContainerName:  pulumi.String(service.name),
					ContainerPort:  pulumi.Int(service.servCfg.ContainerPort),
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

	serviceId := pulumi.Sprintf("service/%s/%s", clusterName, serviceName)
	scalingTarget, err := appautoscaling.NewTarget(
		service.ctx,
		fmt.Sprintf("%s-scaling-target", serviceName),
		&appautoscaling.TargetArgs{
			ResourceId:        serviceId,
			MinCapacity:       pulumi.Int(service.servCfg.MinCount),
			MaxCapacity:       pulumi.Int(service.servCfg.MaxCount),
			ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
			ServiceNamespace:  pulumi.String("ecs"),
		})
	if err != nil {
		log.Fatal(err)
	}

	_, err = appautoscaling.NewPolicy(
		service.ctx,
		fmt.Sprintf("%s-scaling-policy", serviceName),
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
