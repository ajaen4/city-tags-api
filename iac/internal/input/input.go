package input

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Input struct {
	Ctx          *pulumi.Context
	ServicesCfg  map[string]*ServiceCfg
	FunctionsCfg map[string]*FunctionCfg
}

type GCPCfg struct {
	Project string
	Region  string
}

var gcpCfg GCPCfg

func Load(ctx *pulumi.Context) *Input {
	cfg := config.New(ctx, "")

	var servicesCfg map[string]*ServiceCfg
	if err := cfg.TryObject("services", &servicesCfg); err != nil {
		servicesCfg = map[string]*ServiceCfg{}
	}

	var funcsCfg map[string]*FunctionCfg
	if err := cfg.TryObject("functions", &funcsCfg); err != nil {
		funcsCfg = map[string]*FunctionCfg{}
	}

	gcp := config.New(ctx, "gcp")
	gcpCfg = GCPCfg{
		Region:  gcp.Require("region"),
		Project: gcp.Require("project"),
	}

	return &Input{
		Ctx:          ctx,
		ServicesCfg:  servicesCfg,
		FunctionsCfg: funcsCfg,
	}
}

func GetProject() string {
	return gcpCfg.Project
}

func GetRegion() string {
	return gcpCfg.Region
}
