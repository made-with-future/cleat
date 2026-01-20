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
	s.SetReturnToUI(true)
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
