package input

type FunctionCfg struct {
	ImgCfg       ImgCfg   `json:"image"`
	BuildVersion string   `json:"build_version"`
	Memory       int      `json:"memory"`
	EnvVars      []EnvVar `json:"env_vars"`
	Entrypoint   []string `json:"entrypoint"`
}
