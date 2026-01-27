package task

import (
	"fmt"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// NpmBuild runs npm build scripts
type NpmBuild struct {
	BaseTask
	Service *config.ServiceConfig
	Npm     *config.NpmConfig
}

func NewNpmBuild(svc *config.ServiceConfig, npm *config.NpmConfig) *NpmBuild {
	name := "npm:build"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("npm:build:%s", svc.Name)
	}
	return &NpmBuild{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Run NPM build scripts",
			TaskDeps:        []string{"docker:build"}, // Ensure containers are built first
		},
		Service: svc,
		Npm:     npm,
	}
}

func (t *NpmBuild) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Npm != nil && len(t.scriptsToRun()) > 0
}

func (t *NpmBuild) Run(cfg *config.Config, exec executor.Executor) error {
	scripts := t.scriptsToRun()
	if len(scripts) == 0 {
		return nil
	}

	fmt.Printf("==> Running NPM build scripts for service '%s'\n", t.Service.Name)

	for _, script := range scripts {
		fmt.Printf("--> Running npm run %s %s\n", script, modeText(cfg, t.Service, t.Npm))
		cmds := npmScriptCommands(cfg, t.Service, t.Npm, script)
		var err error
		if cfg.Docker && t.Service.IsDocker() && t.Npm.Service != "" {
			err = exec.Run(cmds[0][0], cmds[0][1:]...)
		} else {
			err = exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *NpmBuild) Commands(cfg *config.Config) [][]string {
	var cmds [][]string
	for _, script := range t.scriptsToRun() {
		cmds = append(cmds, npmScriptCommands(cfg, t.Service, t.Npm, script)...)
	}
	return cmds
}

func (t *NpmBuild) scriptsToRun() []string {
	if t.Npm == nil {
		return nil
	}
	// 1. Look for exact match for "build"
	for _, s := range t.Npm.Scripts {
		if s == "build" {
			return []string{"build"}
		}
	}

	// 2. Look for scripts containing "build"
	var buildScripts []string
	for _, s := range t.Npm.Scripts {
		if strings.Contains(strings.ToLower(s), "build") {
			buildScripts = append(buildScripts, s)
		}
	}
	return buildScripts
}

func modeText(cfg *config.Config, svc *config.ServiceConfig, npm *config.NpmConfig) string {
	if cfg.Docker && svc.IsDocker() && npm != nil && npm.Service != "" {
		return fmt.Sprintf("via Docker (%s service)", npm.Service)
	}
	return "locally"
}

// NpmRun runs a single npm script (used by TUI for individual script execution)
type NpmRun struct {
	BaseTask
	Service *config.ServiceConfig
	Npm     *config.NpmConfig
	Script  string
}

func NewNpmRun(svc *config.ServiceConfig, npm *config.NpmConfig, script string) *NpmRun {
	name := fmt.Sprintf("npm:run:%s", script)
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("npm:run:%s:%s", svc.Name, script)
	}
	return &NpmRun{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: fmt.Sprintf("Run npm script: %s", script),
			TaskDeps:        nil,
		},
		Service: svc,
		Npm:     npm,
		Script:  script,
	}
}

func (t *NpmRun) ShouldRun(cfg *config.Config) bool {
	return true // If we created this task, we want to run it
}

func (t *NpmRun) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("--> Running npm run %s %s\n", t.Script, modeText(cfg, t.Service, t.Npm))
	cmds := t.Commands(cfg)
	if cfg.Docker && t.Service.IsDocker() && t.Npm.Service != "" {
		return exec.Run(cmds[0][0], cmds[0][1:]...)
	}
	return exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *NpmRun) Commands(cfg *config.Config) [][]string {
	return npmScriptCommands(cfg, t.Service, t.Npm, t.Script)
}

// NpmStart runs npm start for the run command
type NpmStart struct {
	BaseTask
	Service *config.ServiceConfig
	Npm     *config.NpmConfig
}

func NewNpmStart(svc *config.ServiceConfig, npm *config.NpmConfig) *NpmStart {
	name := "npm:start"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("npm:start:%s", svc.Name)
	}
	return &NpmStart{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Start NPM development server",
			TaskDeps:        nil,
		},
		Service: svc,
		Npm:     npm,
	}
}

func (t *NpmStart) ShouldRun(cfg *config.Config) bool {
	if t.Service == nil || t.Npm == nil {
		return false
	}
	// Run if it's explicitly enabled or we're not using docker globally
	return len(t.Npm.Scripts) > 0
}

func (t *NpmStart) Run(cfg *config.Config, exec executor.Executor) error {
	if cfg.Docker && t.Service.IsDocker() && t.Npm.Service != "" {
		fmt.Printf("==> Running frontend (NPM) for service '%s' via Docker (%s service)\n", t.Service.Name, t.Npm.Service)
	} else {
		fmt.Printf("==> Running frontend (NPM) for service '%s' locally\n", t.Service.Name)
	}
	cmds := t.Commands(cfg)
	if cfg.Docker && t.Service.IsDocker() && t.Npm.Service != "" {
		return exec.Run(cmds[0][0], cmds[0][1:]...)
	}
	return exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *NpmStart) Commands(cfg *config.Config) [][]string {
	if cfg.Docker && t.Service.IsDocker() && t.Npm.Service != "" {
		return [][]string{{"docker", "--log-level", "error", "compose", "run", "--rm", t.Npm.Service, "npm", "run", "start"}}
	}

	return [][]string{{"npm", "run", "start"}}
}

// NpmInstall runs npm install
type NpmInstall struct {
	BaseTask
	Service *config.ServiceConfig
	Npm     *config.NpmConfig
}

func NewNpmInstall(svc *config.ServiceConfig, npm *config.NpmConfig) *NpmInstall {
	name := "npm:install"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("npm:install:%s", svc.Name)
	}
	return &NpmInstall{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Run npm install",
			TaskDeps:        nil,
		},
		Service: svc,
		Npm:     npm,
	}
}

func (t *NpmInstall) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Npm != nil
}

func (t *NpmInstall) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("--> Running npm install %s\n", modeText(cfg, t.Service, t.Npm))
	cmds := t.Commands(cfg)
	if cfg.Docker && t.Service.IsDocker() && t.Npm.Service != "" {
		return exec.Run(cmds[0][0], cmds[0][1:]...)
	}
	return exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *NpmInstall) Commands(cfg *config.Config) [][]string {
	if cfg.Docker && t.Service.IsDocker() && t.Npm.Service != "" {
		return [][]string{{"docker", "--log-level", "error", "compose", "run", "--rm", t.Npm.Service, "npm", "install"}}
	}

	return [][]string{{"npm", "install"}}
}

// npmScriptCommands is a helper for building npm script commands
func npmScriptCommands(cfg *config.Config, svc *config.ServiceConfig, npm *config.NpmConfig, script string) [][]string {
	if cfg.Docker && svc.IsDocker() && npm != nil && npm.Service != "" {
		return [][]string{{"docker", "--log-level", "error", "compose", "run", "--rm", npm.Service, "npm", "run", script}}
	}

	return [][]string{{"npm", "run", script}}
}
