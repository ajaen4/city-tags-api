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

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ImgCfg struct {
	Dockerfile string `json:"dockerfile"`
	Context    string `json:"context"`
}

type ServiceCfg struct {
	Project       string
	Region        string
	ImgCfg        ImgCfg   `json:"image"`
	HostedZoneId  string   `json:"hostedZoneId"`
	DomainName    string   `json:"domainName"`
	BuildVersion  string   `json:"build_version"`
	Cpu           int      `json:"cpu"`
	Memory        string   `json:"memory"`
	MinCount      int      `json:"min_count"`
	MaxCount      int      `json:"max_count"`
	LbPort        int      `json:"lb_port"`
	ContainerPort int      `json:"container_port"`
	EnvVars       []EnvVar `json:"env_vars"`
	Entrypoint    []string `json:"entrypoint"`
}

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

	gcpCfg := config.New(ctx, "gcp")
	region := gcpCfg.Require("region")
	project := gcpCfg.Require("project")
	for _, serviceCfg := range servicesCfg {
		serviceCfg.Region = region
		serviceCfg.Project = project
	}
	for _, funcCfg := range funcsCfg {
		funcCfg.Region = region
		funcCfg.Project = project
	}

	return &Input{
		Ctx:          ctx,
		ServicesCfg:  servicesCfg,
		FunctionsCfg: funcsCfg,
	}
}

type FunctionCfg struct {
	Project      string
	Region       string
	ImgCfg       ImgCfg   `json:"image"`
	BuildVersion string   `json:"build_version"`
	Memory       int      `json:"memory"`
	EnvVars      []EnvVar `json:"env_vars"`
	Entrypoint   []string `json:"entrypoint"`
}
