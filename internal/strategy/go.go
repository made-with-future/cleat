package strategy

import (
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/task"
)

// GoProvider handles Go related commands
type GoProvider struct{}

func (p *GoProvider) CanHandle(command string) bool {
	if !strings.HasPrefix(command, "go ") {
		return false
	}
	act := strings.TrimPrefix(command, "go ")
	if strings.HasPrefix(act, "build") || strings.HasPrefix(act, "test") || strings.HasPrefix(act, "fmt") || strings.HasPrefix(act, "vet") || strings.HasPrefix(act, "mod tidy") || strings.HasPrefix(act, "generate") || strings.HasPrefix(act, "run") || strings.HasPrefix(act, "coverage") || strings.HasPrefix(act, "install") {
		return true
	}
	return false
}

func (p *GoProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}
	// command forms:
	//  - "go build"
	//  - "go build:<svc>"
	//  - "go mod tidy" (normalized in action as "mod-tidy")
	rem := strings.TrimPrefix(command, "go ")
	var svcName string
	if idx := strings.Index(rem, ":"); idx != -1 {
		svcName = rem[idx+1:]
		rem = rem[:idx]
	}
	action := strings.TrimSpace(rem)
	if action == "mod tidy" {
		action = "mod-tidy"
	}

	var targetSvc *config.ServiceConfig
	var goMod *config.GoConfig
	if svcName != "" {
		for i := range sess.Config.Services {
			if sess.Config.Services[i].Name == svcName {
				targetSvc = &sess.Config.Services[i]
				break
			}
		}
		if targetSvc != nil {
			for j := range targetSvc.Modules {
				if targetSvc.Modules[j].Go != nil {
					goMod = targetSvc.Modules[j].Go
					break
				}
			}
		}
	} else {
		for i := range sess.Config.Services {
			for j := range sess.Config.Services[i].Modules {
				if sess.Config.Services[i].Modules[j].Go != nil {
					targetSvc = &sess.Config.Services[i]
					goMod = sess.Config.Services[i].Modules[j].Go
					break
				}
			}
			if targetSvc != nil {
				break
			}
		}
	}

	if targetSvc == nil || goMod == nil {
		return nil
	}

	// handle coverage as a composed set of tasks: test with coverage + show
	if action == "coverage" {
		return NewBaseStrategy("go:coverage", []task.Task{
			task.NewGoAction(targetSvc, goMod, "test-coverage"),
			task.NewGoAction(targetSvc, goMod, "coverage-report"),
		})
	}

	if action == "install" {
		return NewBaseStrategy("go:install", []task.Task{
			task.NewGoInstall(targetSvc, goMod),
		})
	}

	return NewBaseStrategy("go:"+action, []task.Task{
		task.NewGoAction(targetSvc, goMod, action),
	})
}
