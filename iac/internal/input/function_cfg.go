package input

type FunctionCfg struct {
	ImgCfg       ImgCfg   `json:"image"`
	BuildVersion string   `json:"build_version"`
	Memory       int      `json:"memory"`
	ScheduleExp  string   `json:"schedule_exp"`
	EnvVars      []EnvVar `json:"env_vars"`
	Entrypoint   []string `json:"entrypoint"`
}
