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

	// We only want to run docker down once.
	// We prefer the root one if enabled, otherwise the first service-level one we find.
	dockerAdded := false
	if cfg != nil && cfg.Docker {
		tasks = append(tasks, task.NewDockerDown(nil))
		dockerAdded = true
	}

	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() && !dockerAdded {
				tasks = append(tasks, task.NewDockerDown(svc))
				dockerAdded = true
			}
		}
	}

	// If no docker task was added, add the default one (it won't run if cfg.Docker is false)
	if !dockerAdded {
		tasks = append(tasks, task.NewDockerDown(nil))
	}

	return NewBaseStrategy("docker down", tasks)
}

func NewDockerRemoveOrphansStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task

	// We only want to run docker remove-orphans once.
	dockerAdded := false
	if cfg != nil && cfg.Docker {
		tasks = append(tasks, task.NewDockerRemoveOrphans(nil))
		dockerAdded = true
	}

	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() && !dockerAdded {
				tasks = append(tasks, task.NewDockerRemoveOrphans(svc))
				dockerAdded = true
			}
		}
	}

	if !dockerAdded {
		tasks = append(tasks, task.NewDockerRemoveOrphans(nil))
	}

	return NewBaseStrategy("docker remove-orphans", tasks)
}

func NewDockerRebuildStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task

	// We only want to run docker rebuild once.
	dockerAdded := false
	if cfg != nil && cfg.Docker {
		tasks = append(tasks, task.NewDockerRebuild(nil))
		dockerAdded = true
	}

	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() && !dockerAdded {
				tasks = append(tasks, task.NewDockerRebuild(svc))
				dockerAdded = true
			}
		}
	}

	if !dockerAdded {
		tasks = append(tasks, task.NewDockerRebuild(nil))
	}

	return NewBaseStrategy("docker rebuild", tasks)
}
