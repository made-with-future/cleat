package strategy

import (
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func GetNpmInstallStrategy(command string, cfg *config.Config) Strategy {
	svcName := ""
	if strings.Contains(command, ":") {
		svcName = strings.Split(command, ":")[1]
	}

	var targetSvc *config.ServiceConfig
	if svcName != "" {
		for i := range cfg.Services {
			if cfg.Services[i].Name == svcName {
				targetSvc = &cfg.Services[i]
				break
			}
		}
	} else {
		// Default to first service with NPM
		for i := range cfg.Services {
			for j := range cfg.Services[i].Modules {
				if cfg.Services[i].Modules[j].Npm != nil {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc != nil {
				break
			}
		}
	}

	if targetSvc == nil {
		return nil
	}

	var npmMod *config.NpmConfig
	for i := range targetSvc.Modules {
		if targetSvc.Modules[i].Npm != nil {
			npmMod = targetSvc.Modules[i].Npm
			break
		}
	}

	return NewBaseStrategy("npm:install", []task.Task{
		task.NewNpmInstall(targetSvc, npmMod),
	})
}