package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("run", NewRunStrategy)
}

// NewRunStrategy creates the run command strategy
func NewRunStrategy(cfg *config.Config) Strategy {
	var tasks []task.Task

	// We only want to run docker compose up once.
	dockerAdded := false
	if cfg != nil && cfg.Docker {
		tasks = append(tasks, task.NewDockerUp(nil))
		dockerAdded = true
	}

	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() && !dockerAdded {
				tasks = append(tasks, task.NewDockerUp(svc))
				dockerAdded = true
			}
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Python != nil && !cfg.Docker {
					tasks = append(tasks, task.NewDjangoRunServer(svc))
				}
				if mod.Npm != nil && !cfg.Docker {
					// We'll use NewNpmRun for start if it's standardized
					for _, s := range mod.Npm.Scripts {
						if s == "start" {
							tasks = append(tasks, task.NewNpmRun(svc, mod.Npm, "start"))
						}
					}
				}
			}
		}
	}

	if !dockerAdded {
		tasks = append(tasks, task.NewDockerUp(nil))
	}

	return NewBaseStrategy("run", tasks)
}

// NewNpmScriptStrategy creates a strategy for running a single npm script
func NewNpmScriptStrategy(svc *config.ServiceConfig, npm *config.NpmConfig, script string) Strategy {
	return NewBaseStrategy("npm:"+script, []task.Task{
		task.NewNpmRun(svc, npm, script),
	})
}