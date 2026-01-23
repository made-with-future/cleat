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
	tasks = append(tasks, task.NewDockerUp(nil))

	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			if svc.IsDocker() {
				tasks = append(tasks, task.NewDockerUp(svc))
			}
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Python != nil {
					tasks = append(tasks, task.NewDjangoRunServer(svc, mod.Python))
				}
				if mod.Npm != nil {
					tasks = append(tasks, task.NewNpmStart(svc, mod.Npm))
				}
			}
		}
	}

	return NewBaseStrategy("run", tasks)
}

// NewNpmScriptStrategy creates a strategy for running a single npm script
func NewNpmScriptStrategy(svc *config.ServiceConfig, npm *config.NpmConfig, script string) Strategy {
	return NewBaseStrategy("npm:"+script, []task.Task{
		task.NewNpmRun(svc, npm, script),
	})
}
