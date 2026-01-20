package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("gcp activate", func(cfg *config.Config) Strategy {
		return NewGCPActivateStrategy()
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
