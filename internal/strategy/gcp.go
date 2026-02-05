package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("gcp activate", func(cfg *config.Config) Strategy {
		return NewGCPActivateStrategy()
	})
	Register("gcp init", func(cfg *config.Config) Strategy {
		return NewGCPInitStrategy()
	})
	Register("gcp set-config", func(cfg *config.Config) Strategy {
		return NewGCPSetConfigStrategy()
	})
	Register("gcp adc-login", func(cfg *config.Config) Strategy {
		return NewGCPADCLoginStrategy()
	})
	Register("gcp adc-impersonate-login", func(cfg *config.Config) Strategy {
		return NewGCPADCImpersonateLoginStrategy()
	})
	Register("gcp console", func(cfg *config.Config) Strategy {
		return NewGCPConsoleStrategy()
	})
}

func NewGCPActivateStrategy() Strategy {
	return NewBaseStrategy("gcp:activate", []task.Task{
		task.NewGCPActivate(),
	})
}

func NewGCPInitStrategy() Strategy {
	return NewBaseStrategy("gcp:init", []task.Task{
		task.NewGCPInit(),
	})
}

func NewGCPSetConfigStrategy() Strategy {
	return NewBaseStrategy("gcp:set-config", []task.Task{
		task.NewGCPSetConfig(),
	})
}

func NewGCPADCLoginStrategy() Strategy {
	return NewBaseStrategy("gcp:adc-login", []task.Task{
		task.NewGCPADCLogin(),
	})
}

func NewGCPADCImpersonateLoginStrategy() Strategy {
	return NewBaseStrategy("gcp:adc-impersonate-login", []task.Task{
		task.NewGCPAdcImpersonateLogin(),
	})
}

func NewGCPConsoleStrategy() Strategy {
	return NewBaseStrategy("gcp:console", []task.Task{
		task.NewGCPConsole(),
	})
}

func NewGCPAppEngineDeployStrategy(appYaml string) Strategy {
	return NewBaseStrategy("gcp:app-engine-deploy", []task.Task{
		task.NewGCPActivate(),
		task.NewGCPAppEngineDeploy(appYaml),
	})
}

func NewGCPAppEnginePromoteStrategy(service string) Strategy {
	return NewBaseStrategy("gcp:app-engine-promote", []task.Task{
		task.NewGCPActivate(),
		task.NewGCPAppEnginePromote(service),
	})
}