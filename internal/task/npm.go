package task

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// NpmBuild runs npm build scripts
type NpmBuild struct{ BaseTask }

func NewNpmBuild() *NpmBuild {
	return &NpmBuild{
		BaseTask: BaseTask{
			TaskName:        "npm:build",
			TaskDescription: "Run NPM build scripts",
			TaskDeps:        []string{"docker:build"}, // Ensure containers are built first
		},
	}
}

func (t *NpmBuild) ShouldRun(cfg *config.Config) bool {
	return len(cfg.Npm.Scripts) > 0
}

func (t *NpmBuild) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Running NPM build scripts")

	for _, script := range cfg.Npm.Scripts {
		fmt.Printf("--> Running npm run %s %s\n", script, modeText(cfg))
		cmds := npmScriptCommands(cfg, script)
		if err := exec.Run(cmds[0][0], cmds[0][1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *NpmBuild) Commands(cfg *config.Config) [][]string {
	var cmds [][]string
	for _, script := range cfg.Npm.Scripts {
		cmds = append(cmds, npmScriptCommands(cfg, script)...)
	}
	return cmds
}

func modeText(cfg *config.Config) string {
	if cfg.Docker {
		return fmt.Sprintf("via Docker (%s service)", cfg.Npm.Service)
	}
	return "locally"
}

// NpmRun runs a single npm script (used by TUI for individual script execution)
type NpmRun struct {
	BaseTask
	Script string
}

func NewNpmRun(script string) *NpmRun {
	return &NpmRun{
		BaseTask: BaseTask{
			TaskName:        fmt.Sprintf("npm:run:%s", script),
			TaskDescription: fmt.Sprintf("Run npm script: %s", script),
			TaskDeps:        nil,
		},
		Script: script,
	}
}

func (t *NpmRun) ShouldRun(cfg *config.Config) bool {
	return true // If we created this task, we want to run it
}

func (t *NpmRun) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("--> Running npm run %s %s\n", t.Script, modeText(cfg))
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *NpmRun) Commands(cfg *config.Config) [][]string {
	return npmScriptCommands(cfg, t.Script)
}

// NpmStart runs npm start for the run command
type NpmStart struct{ BaseTask }

func NewNpmStart() *NpmStart {
	return &NpmStart{
		BaseTask: BaseTask{
			TaskName:        "npm:start",
			TaskDescription: "Start NPM development server",
			TaskDeps:        nil,
		},
	}
}

func (t *NpmStart) ShouldRun(cfg *config.Config) bool {
	return len(cfg.Npm.Scripts) > 0 && !cfg.Docker && !cfg.Python.Django
}

func (t *NpmStart) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Running frontend (NPM) locally")
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *NpmStart) Commands(cfg *config.Config) [][]string {
	args := []string{"start"}
	if _, err := os.Stat("frontend/package.json"); err == nil {
		args = append([]string{"--prefix", "frontend"}, args...)
	}
	return [][]string{append([]string{"npm"}, args...)}
}

// npmScriptCommands is a helper for building npm script commands
func npmScriptCommands(cfg *config.Config, script string) [][]string {
	if cfg.Docker {
		return [][]string{{"docker", "compose", "run", "--rm", cfg.Npm.Service, "npm", "run", script}}
	}

	args := []string{"run", script}
	if _, err := os.Stat("frontend/package.json"); err == nil {
		args = append([]string{"--prefix", "frontend"}, args...)
	}
	return [][]string{append([]string{"npm"}, args...)}
}
