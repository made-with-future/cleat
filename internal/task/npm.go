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
		if err := runNpmScript(cfg, script, exec); err != nil {
			return err
		}
	}
	return nil
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
	return runNpmScript(cfg, t.Script, exec)
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
	return len(cfg.Npm.Scripts) > 0 && !cfg.Docker && !cfg.Django
}

func (t *NpmStart) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Running frontend (NPM) locally")
	args := []string{"start"}
	if _, err := os.Stat("frontend/package.json"); err == nil {
		args = append([]string{"--prefix", "frontend"}, args...)
	}
	return exec.Run("npm", args...)
}

// runNpmScript is a helper for running npm scripts
func runNpmScript(cfg *config.Config, script string, exec executor.Executor) error {
	if cfg.Docker {
		fmt.Printf("--> Running npm run %s via Docker (%s service)\n", script, cfg.Npm.Service)
		return exec.Run("docker", "compose", "run", "--rm", cfg.Npm.Service, "npm", "run", script)
	}

	fmt.Printf("--> Running npm run %s locally\n", script)
	args := []string{"run", script}
	if _, err := os.Stat("frontend/package.json"); err == nil {
		args = append([]string{"--prefix", "frontend"}, args...)
	}
	return exec.Run("npm", args...)
}
