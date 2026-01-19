package task

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// DockerBuild builds Docker containers
type DockerBuild struct{ BaseTask }

func NewDockerBuild() *DockerBuild {
	return &DockerBuild{
		BaseTask: BaseTask{
			TaskName:        "docker:build",
			TaskDescription: "Build Docker containers",
			TaskDeps:        nil,
		},
	}
}

func (t *DockerBuild) ShouldRun(cfg *config.Config) bool {
	return cfg.Docker
}

func (t *DockerBuild) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Building Docker containers")
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DockerBuild) Commands(cfg *config.Config) [][]string {
	return [][]string{{"docker", "compose", "build"}}
}

// DockerUp starts Docker containers
type DockerUp struct{ BaseTask }

func NewDockerUp() *DockerUp {
	return &DockerUp{
		BaseTask: BaseTask{
			TaskName:        "docker:up",
			TaskDescription: "Start Docker containers",
			TaskDeps:        nil,
		},
	}
}

func (t *DockerUp) ShouldRun(cfg *config.Config) bool {
	return cfg.Docker
}

func (t *DockerUp) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Running project via Docker")

	// 1Password integration
	if _, err := os.Stat(".env/dev.env"); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DockerUp) Commands(cfg *config.Config) [][]string {
	cmdName := "docker"
	args := []string{"compose", "up", "--remove-orphans"}

	// 1Password integration
	if _, err := os.Stat(".env/dev.env"); err == nil {
		args = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, args...)
		cmdName = "op"
	}

	return [][]string{append([]string{cmdName}, args...)}
}

// DockerDown stops Docker containers
type DockerDown struct{ BaseTask }

func NewDockerDown() *DockerDown {
	return &DockerDown{
		BaseTask: BaseTask{
			TaskName:        "docker:down",
			TaskDescription: "Stop Docker containers",
			TaskDeps:        nil,
		},
	}
}

func (t *DockerDown) ShouldRun(cfg *config.Config) bool {
	return cfg.Docker
}

func (t *DockerDown) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Stopping Docker containers (all profiles)")

	// 1Password integration
	if _, err := os.Stat(".env/dev.env"); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DockerDown) Commands(cfg *config.Config) [][]string {
	cmdName := "docker"
	args := []string{"compose", "--profile", "*", "down", "--remove-orphans"}

	// 1Password integration
	if _, err := os.Stat(".env/dev.env"); err == nil {
		args = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, args...)
		cmdName = "op"
	}

	return [][]string{append([]string{cmdName}, args...)}
}

// DockerRebuild stops all containers, removes images/volumes, and rebuilds without cache
type DockerRebuild struct{ BaseTask }

func NewDockerRebuild() *DockerRebuild {
	return &DockerRebuild{
		BaseTask: BaseTask{
			TaskName:        "docker:rebuild",
			TaskDescription: "Stop, remove all and rebuild Docker containers (all profiles, no cache)",
			TaskDeps:        nil,
		},
	}
}

func (t *DockerRebuild) ShouldRun(cfg *config.Config) bool {
	return cfg.Docker
}

func (t *DockerRebuild) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Rebuilding Docker containers (all profiles, no cache)")

	// 1. Down with --rmi all --volumes
	fmt.Println("--> Cleaning up: stopping containers and removing images/volumes")
	if _, err := os.Stat(".env/dev.env"); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
	}

	cmds := t.Commands(cfg)
	if err := exec.Run(cmds[0][0], cmds[0][1:]...); err != nil {
		return err
	}

	// 2. Build with --no-cache
	fmt.Println("--> Rebuilding: build without cache")
	return exec.Run(cmds[1][0], cmds[1][1:]...)
}

func (t *DockerRebuild) Commands(cfg *config.Config) [][]string {
	// 1. Down
	downCmd := "docker"
	downArgs := []string{"compose", "--profile", "*", "down", "--remove-orphans", "--rmi", "all", "--volumes"}

	if _, err := os.Stat(".env/dev.env"); err == nil {
		downArgs = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, downArgs...)
		downCmd = "op"
	}

	// 2. Build
	buildCmd := "docker"
	buildArgs := []string{"compose", "--profile", "*", "build", "--no-cache"}

	if _, err := os.Stat(".env/dev.env"); err == nil {
		buildArgs = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, buildArgs...)
		buildCmd = "op"
	}

	return [][]string{
		append([]string{downCmd}, downArgs...),
		append([]string{buildCmd}, buildArgs...),
	}
}
