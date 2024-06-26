package containers

import (
	"fmt"
	"log"

	"city-tags-api-iac/internal/config"
	"city-tags-api-iac/internal/types"

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

	execRoleName := fmt.Sprintf("execution-role-%s", servs.cfg.Ctx.Stack())
	execRole, err := iam.NewRole(
		servs.cfg.Ctx,
		execRoleName,
		&iam.RoleArgs{
			Name:             pulumi.String(execRoleName),
			AssumeRolePolicy: pulumi.String(assumeRPol),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	execPolName := fmt.Sprintf("execution-role-policy-%s", servs.cfg.Ctx.Stack())
	_, err = iam.NewRolePolicy(
		servs.cfg.Ctx,
		execPolName,
		&iam.RolePolicyArgs{
			Name:   pulumi.String(execPolName),
			Role:   execRole.ID(),
			Policy: pulumi.String(execPol),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	taskRoleName := fmt.Sprintf("task-role-%s", servs.cfg.Ctx.Stack())
	taskRole, err := iam.NewRole(
		servs.cfg.Ctx,
		taskRoleName,
		&iam.RoleArgs{
			Name:             pulumi.String(taskRoleName),
			AssumeRolePolicy: pulumi.String(assumeRPol),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	taskPolicyName := fmt.Sprintf("task-role-policy-%s", servs.cfg.Ctx.Stack())
	_, err = iam.NewRolePolicy(
		servs.cfg.Ctx,
		taskPolicyName,
		&iam.RolePolicyArgs{
			Name:   pulumi.String(taskPolicyName),
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
