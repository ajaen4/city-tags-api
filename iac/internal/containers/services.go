package containers

import (
	"city-tags-api-iac/internal/config"
	"city-tags-api-iac/internal/types"
	"log"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Services struct {
	cfg *config.Config
}

func NewServices(cfg *config.Config) *Services {
	return &Services{
		cfg: cfg,
	}
}

func (servs *Services) Deploy() {
	roles := servs.createRoles()
	for servName, servCfg := range servs.cfg.ServicesCfg {
		service := NewService(servs.cfg.Ctx, servName, servCfg, roles)
		service.deploy()
	}
}

func (servs *Services) createRoles() map[string]*iam.Role {
	assumeRPol := types.ValidateJSON(`{
		"Version": "2008-10-17",
		"Statement": [
			{
				"Action": "sts:AssumeRole",
				"Principal": {"Service": "ecs-tasks.amazonaws.com"},
				"Effect": "Allow"
			}
		]
	}`)

	execPol := types.ValidateJSON(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ecs:StartTask",
					"ecs:StopTask",
					"ecs:DescribeTasks",
					"ecr:GetAuthorizationToken",
					"ecr:BatchCheckLayerAvailability",
					"ecr:GetDownloadUrlForLayer",
					"ecr:BatchGetImage",
					"logs:CreateLogStream",
					"logs:PutLogEvents",
					"elasticfilesystem:ClientMount",
					"elasticfilesystem:ClientWrite",
					"elasticfilesystem:ClientRootAccess",
					"elasticfilesystem:DescribeFileSystems"
				],
				"Resource": "*"
			}
		]
	}`)

	taskPolicy := types.ValidateJSON(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"elasticloadbalancing:Describe*",
					"elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
					"elasticloadbalancing:RegisterInstancesWithLoadBalancer",
					"ec2:Describe*",
					"ec2:AuthorizeSecurityGroupIngress",
					"elasticloadbalancing:RegisterTargets",
					"elasticloadbalancing:DeregisterTargets",
					"elasticfilesystem:ClientMount",
					"elasticfilesystem:ClientWrite",
					"elasticfilesystem:ClientRootAccess",
					"elasticfilesystem:DescribeFileSystems",
					"ssm:*"
				],
				"Resource": "*"
			}
		]
	}`)

	execRole, err := iam.NewRole(
		servs.cfg.Ctx,
		"execution_role",
		&iam.RoleArgs{
			Name:             pulumi.String("execution-role"),
			AssumeRolePolicy: pulumi.String(assumeRPol),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = iam.NewRolePolicy(
		servs.cfg.Ctx,
		"execution-role-policy",
		&iam.RolePolicyArgs{
			Name:   pulumi.String("execution-role-policy"),
			Role:   execRole.ID(),
			Policy: pulumi.String(execPol),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	taskRole, err := iam.NewRole(
		servs.cfg.Ctx,
		"task_role",
		&iam.RoleArgs{
			Name:             pulumi.String("task-role"),
			AssumeRolePolicy: pulumi.String(assumeRPol),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = iam.NewRolePolicy(
		servs.cfg.Ctx,
		"task-role-policy",
		&iam.RolePolicyArgs{
			Name:   pulumi.String("task-role-policy"),
			Role:   taskRole.ID(),
			Policy: pulumi.String(taskPolicy),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return map[string]*iam.Role{
		"exec_role": execRole,
		"task_role": taskRole,
	}
}
