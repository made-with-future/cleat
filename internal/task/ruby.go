package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

type RubyAction struct {
	BaseTask
	Service *config.ServiceConfig
	RubyCfg *config.RubyConfig
	Action  string
}

func NewRubyAction(svc *config.ServiceConfig, r *config.RubyConfig, action string) *RubyAction {
	return &RubyAction{
		BaseTask: BaseTask{
			TaskName:        fmt.Sprintf("ruby:%s", action),
			TaskDescription: fmt.Sprintf("Run ruby action: %s", action),
		},
		Service: svc,
		RubyCfg: r,
		Action:  action,
	}
}

func (t *RubyAction) ShouldRun(sess *session.Session) bool {
	return t.RubyCfg != nil && t.RubyCfg.IsEnabled()
}

func (t *RubyAction) Run(sess *session.Session) error {
	desc := fmt.Sprintf("Running 'ruby %s' for service %s", t.Action, t.Service.Name)
	if sess.Config.Docker && t.Service.IsDocker() && t.RubyCfg.RailsService != "" {
		desc += fmt.Sprintf(" via Docker (%s service)", t.RubyCfg.RailsService)
	}
	PrintStep(desc)

	cmd := t.commandArgs(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() && t.RubyCfg.RailsService != "" {
		dir = ""
	}

	if err := sess.Exec.RunWithDir(dir, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("ruby %s failed for service %s: %w", t.Action, t.Service.Name, err)
	}
	return nil
}

func (t *RubyAction) commandArgs(sess *session.Session) []string {
	args := t.argsForAction()
	if sess.Config.Docker && t.Service.IsDocker() && t.RubyCfg.RailsService != "" {
		base := []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.RubyCfg.RailsService}
		if t.RubyCfg.Rails {
			return append(base, append([]string{"bundle", "exec"}, args...)...)
		}
		return append(base, args...)
	}

	// Local execution
	envCmd := detectRubyEnvCommand(t.Service.Dir, sess.Config.SourcePath)
	if t.RubyCfg.Rails {
		return append(envCmd, append([]string{"bundle", "exec"}, args...)...)
	}
	return append(envCmd, args...)
}

func (t *RubyAction) argsForAction() []string {
	switch t.Action {
	case "migrate":
		return []string{"rails", "db:migrate"}
	case "console":
		return []string{"rails", "console"}
	case "server":
		return []string{"rails", "server", "-b", "0.0.0.0"}
	case "assets:precompile":
		return []string{"rails", "assets:precompile"}
	default:
		return []string{t.Action}
	}
}

func (t *RubyAction) Commands(sess *session.Session) [][]string {
	return [][]string{t.commandArgs(sess)}
}

type RubyInstall struct {
	BaseTask
	Service *config.ServiceConfig
	RubyCfg *config.RubyConfig
}

func NewRubyInstall(svc *config.ServiceConfig, r *config.RubyConfig) *RubyInstall {
	return &RubyInstall{
		BaseTask: BaseTask{
			TaskName:        "ruby:install",
			TaskDescription: "Run bundle install",
		},
		Service: svc,
		RubyCfg: r,
	}
}

func (t *RubyInstall) ShouldRun(sess *session.Session) bool {
	return t.RubyCfg != nil && t.RubyCfg.IsEnabled()
}

func (t *RubyInstall) Run(sess *session.Session) error {
	PrintStep(fmt.Sprintf("Running bundle install for service %s", t.Service.Name))
	cmd := t.commandArgs(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() && t.RubyCfg.RailsService != "" {
		dir = ""
	}
	if err := sess.Exec.RunWithDir(dir, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("bundle install failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *RubyInstall) commandArgs(sess *session.Session) []string {
	args := []string{"bundle", "install"}
	if sess.Config.Docker && t.Service.IsDocker() && t.RubyCfg.RailsService != "" {
		return []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.RubyCfg.RailsService, "bundle", "install"}
	}
	return args
}

func (t *RubyInstall) Commands(sess *session.Session) [][]string {
	return [][]string{t.commandArgs(sess)}
}

func detectRubyEnvCommand(dir string, configSourcePath string) []string {
	// 1. Check for asdf
	if _, err := os.Stat(filepath.Join(dir, ".tool-versions")); err == nil {
		return []string{} // asdf usually handles it via shims
	}

	// 2. Check for rbenv
	if _, err := os.Stat(filepath.Join(dir, ".ruby-version")); err == nil {
		return []string{} // rbenv handles it via shims
	}

	// 3. Check for rvm
	if _, err := os.Stat(filepath.Join(dir, ".rvmrc")); err == nil {
		return []string{}
	}

	// Default to just using the environment as is
	return []string{}
}
