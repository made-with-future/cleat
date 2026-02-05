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

type GCPActivateStrategy struct {
	BaseStrategy
}

func NewGCPActivateStrategy() *GCPActivateStrategy {
	return &GCPActivateStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:activate", []task.Task{
			task.NewGCPActivate(),
		}),
	}
}

func (s *GCPActivateStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

type GCPInitStrategy struct {
	BaseStrategy
}

func NewGCPInitStrategy() *GCPInitStrategy {
	s := &GCPInitStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:init", []task.Task{
			task.NewGCPInit(),
			task.NewGCPSetConfig(),
		}),
	}
	return s
}

func (s *GCPInitStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

type GCPSetConfigStrategy struct {
	BaseStrategy
}

func NewGCPSetConfigStrategy() *GCPSetConfigStrategy {
	return &GCPSetConfigStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:set-config", []task.Task{
			task.NewGCPActivate(),
			task.NewGCPSetConfig(),
		}),
	}
}

func (s *GCPSetConfigStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

type GCPADCLoginStrategy struct {
	BaseStrategy
}

func NewGCPADCLoginStrategy() *GCPADCLoginStrategy {
	return &GCPADCLoginStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:adc-login", []task.Task{
			task.NewGCPADCLogin(),
		}),
	}
}

func (s *GCPADCLoginStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

type GCPADCImpersonateLoginStrategy struct {
	BaseStrategy
}

func NewGCPADCImpersonateLoginStrategy() *GCPADCImpersonateLoginStrategy {
	return &GCPADCImpersonateLoginStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:adc-impersonate-login", []task.Task{
			task.NewGCPADCImpersonateLogin(),
		}),
	}
}

func (s *GCPADCImpersonateLoginStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

type GCPAppEngineDeployStrategy struct {
	BaseStrategy
}

func NewGCPAppEngineDeployStrategy(appYaml string) *GCPAppEngineDeployStrategy {
	return &GCPAppEngineDeployStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:app-engine-deploy", []task.Task{
			task.NewGCPActivate(),
			task.NewGCPAppEngineDeploy(appYaml),
		}),
	}
}

func (s *GCPAppEngineDeployStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

type GCPAppEnginePromoteStrategy struct {
	BaseStrategy
}

func NewGCPAppEnginePromoteStrategy(service string) *GCPAppEnginePromoteStrategy {
	return &GCPAppEnginePromoteStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:app-engine-promote", []task.Task{
			task.NewGCPActivate(),
			task.NewGCPAppEnginePromote(service),
		}),
	}
}

func (s *GCPAppEnginePromoteStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}

type GCPConsoleStrategy struct {
	BaseStrategy
}

func NewGCPConsoleStrategy() *GCPConsoleStrategy {
	return &GCPConsoleStrategy{
		BaseStrategy: *NewBaseStrategy("gcp:console", []task.Task{
			task.NewGCPConsole(),
		}),
	}
}

func (s *GCPConsoleStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}
