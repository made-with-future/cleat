package strategy

import (
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("npm install", func(cfg *config.Config) Strategy {
		return NewNpmInstallStrategy(nil, nil)
	})
}

type NpmInstallStrategy struct {
	BaseStrategy
}

func NewNpmInstallStrategy(svc *config.ServiceConfig, npm *config.NpmConfig) *NpmInstallStrategy {
	return &NpmInstallStrategy{
		BaseStrategy: *NewBaseStrategy("npm:install", []task.Task{
			task.NewNpmInstall(svc, npm),
		}),
	}
}

func (s *NpmInstallStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	if s.tasks[0].(*task.NpmInstall).Service == nil {
		// This happens when called from CLI without specific service
		// In that case we run it for all NPM modules? Or just the first one?
		// Requirement says "anytime npm is detected or configured, we need to make sure there's a way to run npm install in that context"

		var tasks []task.Task
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Npm != nil {
					tasks = append(tasks, task.NewNpmInstall(svc, mod.Npm))
				}
			}
		}
		return tasks, nil
	}
	return s.buildExecutionPlan(cfg)
}

func GetNpmInstallStrategy(command string, cfg *config.Config) Strategy {
	if strings.HasPrefix(command, "npm install:") {
		svcName := strings.TrimPrefix(command, "npm install:")
		for i := range cfg.Services {
			if cfg.Services[i].Name == svcName {
				svc := &cfg.Services[i]
				for j := range svc.Modules {
					if svc.Modules[j].Npm != nil {
						return NewNpmInstallStrategy(svc, svc.Modules[j].Npm)
					}
				}
			}
		}
	}
	return nil
}
