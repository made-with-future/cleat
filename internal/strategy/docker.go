package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("docker down", NewDockerDownStrategy)
	Register("docker rebuild", NewDockerRebuildStrategy)
	Register("docker remove-orphans", NewDockerRemoveOrphansStrategy)
}

func NewDockerDownStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task
	tasks = append(tasks, task.NewDockerDown(nil))
	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() {
				tasks = append(tasks, task.NewDockerDown(svc))
			}
		}
	}
	return NewBaseStrategy("docker down", tasks)
}

func NewDockerRemoveOrphansStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task
	tasks = append(tasks, task.NewDockerRemoveOrphans(nil))
	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() {
				tasks = append(tasks, task.NewDockerRemoveOrphans(svc))
			}
		}
	}
	return NewBaseStrategy("docker remove-orphans", tasks)
}

func NewDockerRebuildStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task
	tasks = append(tasks, task.NewDockerRebuild(nil))
	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() {
				tasks = append(tasks, task.NewDockerRebuild(svc))
			}
		}
	}
	return NewBaseStrategy("docker rebuild", tasks)
}
