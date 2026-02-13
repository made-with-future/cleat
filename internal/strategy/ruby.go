package strategy

import (
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/task"
)

// RubyProvider handles service-specific ruby commands
type RubyProvider struct{}

func (p *RubyProvider) CanHandle(command string) bool {
	return strings.HasPrefix(command, "ruby ") || strings.HasPrefix(command, "ruby:")
}

func (p *RubyProvider) GetStrategy(command string, sess *session.Session) Strategy {
	if sess == nil {
		return nil
	}

	// Normalize: "ruby migrate" -> "ruby:migrate"
	normalized := strings.Replace(command, " ", ":", 1)
	parts := strings.Split(normalized, ":")

	if len(parts) == 2 {
		baseCmd := parts[1]
		targetSvc, rubyCfg := p.findRubyService(sess, "")
		if targetSvc != nil && rubyCfg != nil {
			if baseCmd == "install" {
				return NewBaseStrategy("ruby:install", []task.Task{
					task.NewRubyInstall(targetSvc, rubyCfg),
				})
			}
			return NewBaseStrategy("ruby:"+baseCmd, []task.Task{
				task.NewRubyAction(targetSvc, rubyCfg, baseCmd),
			})
		}
	}

	if len(parts) == 3 {
		baseCmd := parts[1]
		svcName := parts[2]
		targetSvc, rubyCfg := p.findRubyService(sess, svcName)
		if targetSvc != nil && rubyCfg != nil {
			if baseCmd == "install" {
				return NewBaseStrategy("ruby:install", []task.Task{
					task.NewRubyInstall(targetSvc, rubyCfg),
				})
			}
			return NewBaseStrategy("ruby:"+baseCmd, []task.Task{
				task.NewRubyAction(targetSvc, rubyCfg, baseCmd),
			})
		}
	}

	return nil
}

func (p *RubyProvider) findRubyService(sess *session.Session, svcName string) (*config.ServiceConfig, *config.RubyConfig) {
	if svcName != "" {
		for i := range sess.Config.Services {
			if sess.Config.Services[i].Name == svcName {
				svc := &sess.Config.Services[i]
				for j := range svc.Modules {
					if svc.Modules[j].Ruby != nil {
						return svc, svc.Modules[j].Ruby
					}
				}
			}
		}
	} else {
		for i := range sess.Config.Services {
			svc := &sess.Config.Services[i]
			for j := range svc.Modules {
				if svc.Modules[j].Ruby != nil {
					return svc, svc.Modules[j].Ruby
				}
			}
		}
	}
	return nil, nil
}
