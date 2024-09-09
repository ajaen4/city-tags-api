package input

type ServiceCfg struct {
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
