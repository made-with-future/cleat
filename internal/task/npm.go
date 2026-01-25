package task

import (
	"fmt"
	"os"
	"path/filepath"

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
	return t.Service != nil && t.Npm != nil && len(t.Npm.Scripts) > 0
}

func (t *NpmBuild) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Running NPM build scripts for service '%s'\n", t.Service.Name)

	for _, script := range t.Npm.Scripts {
		fmt.Printf("--> Running npm run %s %s\n", script, modeText(cfg, t.Service, t.Npm))
		cmds := npmScriptCommands(cfg, t.Service, t.Npm, script)
		if err := exec.Run(cmds[0][0], cmds[0][1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *NpmBuild) Commands(cfg *config.Config) [][]string {
	var cmds [][]string
	for _, script := range t.Npm.Scripts {
		cmds = append(cmds, npmScriptCommands(cfg, t.Service, t.Npm, script)...)
	}
	return cmds
}

func modeText(cfg *config.Config, svc *config.ServiceConfig, npm *config.NpmConfig) string {
	if cfg.Docker && npm != nil && npm.Service != "" {
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
	return exec.Run(cmds[0][0], cmds[0][1:]...)
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
	// Only run if not using docker/django globally (legacy behavior) OR if it's explicitly defined
	return len(t.Npm.Scripts) > 0 && !cfg.Docker
}

func (t *NpmStart) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Running frontend (NPM) for service '%s' locally\n", t.Service.Name)
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *NpmStart) Commands(cfg *config.Config) [][]string {
	args := []string{"run", "start"} // Default to 'npm run start'
	// Check for 'frontend' subdir relative to service dir
	pkgPath := filepath.Join(t.Service.Dir, "package.json")
	if _, err := os.Stat(filepath.Join(t.Service.Dir, "frontend/package.json")); err == nil {
		args = append([]string{"--prefix", filepath.Join(t.Service.Dir, "frontend")}, "start")
	} else if _, err := os.Stat(pkgPath); err == nil {
		if t.Service.Dir != "" && t.Service.Dir != "." {
			args = append([]string{"--prefix", t.Service.Dir}, "start")
		} else {
			args = []string{"run", "start"}
		}
	}
	return [][]string{append([]string{"npm"}, args...)}
}

// npmScriptCommands is a helper for building npm script commands
func npmScriptCommands(cfg *config.Config, svc *config.ServiceConfig, npm *config.NpmConfig, script string) [][]string {
	if cfg.Docker && npm != nil && npm.Service != "" {
		return [][]string{{"docker", "--log-level", "error", "compose", "run", "--rm", npm.Service, "npm", "run", script}}
	}

	args := []string{"run", script}
	if svc != nil {
		if _, err := os.Stat(filepath.Join(svc.Dir, "frontend/package.json")); err == nil {
			args = append([]string{"--prefix", filepath.Join(svc.Dir, "frontend")}, script)
		} else if svc.Dir != "" && svc.Dir != "." {
			args = append([]string{"--prefix", svc.Dir}, script)
		}
	}
	return [][]string{append([]string{"npm"}, args...)}
}
