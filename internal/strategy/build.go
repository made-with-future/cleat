package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("build", NewBuildStrategy)
}

// NewBuildStrategy creates the build command strategy
func NewBuildStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task

	// We only want to run docker build once.
	dockerAdded := false
	if cfg != nil && cfg.Docker {
		tasks = append(tasks, task.NewDockerBuild(nil))
		dockerAdded = true
	}

	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() && !dockerAdded {
				tasks = append(tasks, task.NewDockerBuild(svc))
				dockerAdded = true
			}
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Npm != nil {
					tasks = append(tasks, task.NewNpmBuild(svc, mod.Npm))
				}
				if mod.Python != nil {
					tasks = append(tasks, task.NewDjangoCollectStatic(svc, mod.Python))
				}
			}
		}
	}

	if !dockerAdded {
		tasks = append(tasks, task.NewDockerBuild(nil))
	}

	return NewBaseStrategy("build", tasks)
}
