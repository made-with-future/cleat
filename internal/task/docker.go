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
