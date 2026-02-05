package task

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

type NpmRun struct {
	BaseTask
	Service *config.ServiceConfig
	Npm     *config.NpmConfig
	Script  string
}

func NewNpmRun(svc *config.ServiceConfig, npm *config.NpmConfig, script string) *NpmRun {
	return &NpmRun{
		BaseTask: BaseTask{
			TaskName:        fmt.Sprintf("npm:run:%s", script),
			TaskDescription: fmt.Sprintf("Run NPM script: %s", script),
		},
		Service: svc,
		Npm:     npm,
		Script:  script,
	}
}

func (t *NpmRun) ShouldRun(sess *session.Session) bool {
	return t.Npm != nil && t.Npm.IsEnabled()
}

func (t *NpmRun) Run(sess *session.Session) error {
	fmt.Printf("==> Running NPM script %s for service %s\n", t.Script, t.Service.Name)
	cmds := t.Commands(sess)
	if err := sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("npm run %s failed for service %s: %w", t.Script, t.Service.Name, err)
	}
	return nil
}

func (t *NpmRun) Commands(sess *session.Session) [][]string {
	return [][]string{{"npm", "run", t.Script}}
}

type NpmInstall struct {
	BaseTask
	Service *config.ServiceConfig
	Npm     *config.NpmConfig
}

func NewNpmInstall(svc *config.ServiceConfig, npm *config.NpmConfig) *NpmInstall {
	return &NpmInstall{
		BaseTask: BaseTask{
			TaskName:        "npm:install",
			TaskDescription: "Install NPM dependencies",
		},
		Service: svc,
		Npm:     npm,
	}
}

func (t *NpmInstall) ShouldRun(sess *session.Session) bool {
	return t.Npm != nil && t.Npm.IsEnabled()
}

func (t *NpmInstall) Run(sess *session.Session) error {
	fmt.Printf("==> Installing dependencies for service %s\n", t.Service.Name)
	cmds := t.Commands(sess)
	if err := sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("npm install failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *NpmInstall) Commands(sess *session.Session) [][]string {
	return [][]string{{"npm", "install"}}
}
