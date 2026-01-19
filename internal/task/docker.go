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
	return exec.Run("docker", "compose", "build")
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
	cmdName := "docker"
	args := []string{"compose", "up", "--remove-orphans"}

	// 1Password integration
	if _, err := os.Stat(".env/dev.env"); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
		args = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, args...)
		cmdName = "op"
	}

	return exec.Run(cmdName, args...)
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
	cmdName := "docker"
	args := []string{"compose", "--profile", "*", "down", "--remove-orphans"}

	// 1Password integration
	if _, err := os.Stat(".env/dev.env"); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
		args = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, args...)
		cmdName = "op"
	}

	return exec.Run(cmdName, args...)
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
	downCmd := "docker"
	downArgs := []string{"compose", "--profile", "*", "down", "--remove-orphans", "--rmi", "all", "--volumes"}

	if _, err := os.Stat(".env/dev.env"); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
		downArgs = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, downArgs...)
		downCmd = "op"
	}

	if err := exec.Run(downCmd, downArgs...); err != nil {
		return err
	}

	// 2. Build with --no-cache
	fmt.Println("--> Rebuilding: build without cache")
	buildCmd := "docker"
	buildArgs := []string{"compose", "--profile", "*", "build", "--no-cache"}

	if _, err := os.Stat(".env/dev.env"); err == nil {
		buildArgs = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, buildArgs...)
		buildCmd = "op"
	}

	return exec.Run(buildCmd, buildArgs...)
}
