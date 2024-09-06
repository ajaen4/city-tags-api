package containers

import (
	"city-tags-api-iac/internal/input"
)

type Functions struct {
	cfg *input.Input
}

func NewFunctions(cfg *input.Input) *Functions {
	return &Functions{
		cfg: cfg,
	}
}

func (funcs *Functions) Deploy() {
	for funcName, funcCfg := range funcs.cfg.FunctionsCfg {
		service := NewFunction(funcs.cfg.Ctx, funcName, funcCfg)
		service.deploy()
	}
}
