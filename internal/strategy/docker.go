package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("docker down", NewDockerDownStrategy)
	Register("docker rebuild", NewDockerRebuildStrategy)
}

func NewDockerDownStrategy(cfg *config.Config) Strategy {
	return NewBaseStrategy("docker down", []task.Task{
		task.NewDockerDown(),
	})
}

func NewDockerRebuildStrategy(cfg *config.Config) Strategy {
	return NewBaseStrategy("docker rebuild", []task.Task{
		task.NewDockerRebuild(),
	})
}
