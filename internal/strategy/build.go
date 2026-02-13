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
		}

		// Add NPM build tasks
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Npm != nil {
					for _, s := range mod.Npm.Scripts {
						if s == "build" {
							tasks = append(tasks, task.NewNpmRun(svc, mod.Npm, "build"))
						}
					}
				}
			}
		}

		// Add Go build tasks
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Go != nil {
					tasks = append(tasks, task.NewGoAction(svc, mod.Go, "build"))
				}
			}
		}

		// Add Django collectstatic tasks
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Python != nil {
					tasks = append(tasks, task.NewDjangoCollectStatic(svc))
				}
			}
		}

		// Add Rails asset precompilation tasks
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Ruby != nil && mod.Ruby.Rails {
					tasks = append(tasks, task.NewRubyAction(svc, mod.Ruby, "assets:precompile"))
				}
			}
		}
	}

	if !dockerAdded {
		tasks = append(tasks, task.NewDockerBuild(nil))
	}

	return NewBaseStrategy("build", tasks)
}
