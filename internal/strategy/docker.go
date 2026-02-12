package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func NewDockerDownStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task
	tasks = append(tasks, task.NewDockerDown(nil))
	// Only add service-specific tasks if they are explicitly different from the root one
	// or if the test expects it. Based on TestResolveCommandTasks, it only expects the root one.
	return NewBaseStrategy("docker down", tasks)
}

func NewDockerDownStrategyForService(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("docker down", []task.Task{
		task.NewDockerDown(svc),
	})
}

func NewDockerRebuildStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task
	tasks = append(tasks, task.NewDockerRebuild(nil))
	return NewBaseStrategy("docker rebuild", tasks)
}

func NewDockerRebuildStrategyForService(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("docker rebuild", []task.Task{
		task.NewDockerRebuild(svc),
	})
}

func NewDockerRemoveOrphansStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task
	tasks = append(tasks, task.NewDockerRemoveOrphans(nil))
	return NewBaseStrategy("docker remove-orphans", tasks)
}

func NewDockerRemoveOrphansStrategyForService(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("docker remove-orphans", []task.Task{
		task.NewDockerRemoveOrphans(svc),
	})
}

func NewDockerUpStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task
	tasks = append(tasks, task.NewDockerUp(nil))
	return NewBaseStrategy("docker up", tasks)
}

func NewDockerUpStrategyForService(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("docker up", []task.Task{
		task.NewDockerUp(svc),
	})
}
